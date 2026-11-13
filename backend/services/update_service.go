package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
	"crac_exam_v3/backend/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const officialBankPage = "https://www.crac.org.cn/?p=1287"

var (
	zipLinkRe  = regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+\.zip)["'][^>]*>(.*?)</a>`)
	postDateRe = regexp.MustCompile(`(?is)<span class="post-date">(.*?)</span>`)
)

type remoteBank struct {
	URL          string
	PageURL      string
	Version      string
	Size         int64
	ETag         string
	LastModified string
	Hash         string
	TempPath     string
}

// UpdateService 检查并应用 CRAC 官方题库更新。
type UpdateService struct {
	db *sql.DB
	mu sync.Mutex

	status      *models.BankUpdateStatus
	checking    bool
	checkWait   chan struct{}
	pendingPath string
	pending     remoteBank
}

func NewUpdateService(db *sql.DB) *UpdateService {
	return &UpdateService{db: db}
}

func (s *UpdateService) ServiceName() string { return "UpdateService" }

func boolString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func parseBool(v string, def bool) bool {
	if v == "" {
		return def
	}
	return v == "1"
}

func (s *UpdateService) GetSettings() *models.BankUpdateSettings {
	check, _ := store.GetMeta(s.db, store.MetaBankAutoCheck)
	imp, _ := store.GetMeta(s.db, store.MetaBankAutoImport)
	return &models.BankUpdateSettings{
		AutoCheck:  parseBool(check, true),
		AutoImport: parseBool(imp, false),
	}
}

func (s *UpdateService) SetAutoUpdate(autoCheck, autoImport bool) error {
	if err := store.SetMetas(s.db, map[string]string{
		store.MetaBankAutoCheck:  boolString(autoCheck),
		store.MetaBankAutoImport: boolString(autoImport),
	}); err != nil {
		return err
	}
	s.mu.Lock()
	if s.status == nil {
		s.status = s.statusFromMeta()
	}
	s.status.AutoCheck = autoCheck
	s.status.AutoImport = autoImport
	s.mu.Unlock()
	s.emitStatus()
	return nil
}

func (s *UpdateService) statusFromMeta() *models.BankUpdateStatus {
	meta, _ := store.GetMetaMap(s.db, "bank.")
	st := &models.BankUpdateStatus{
		CurrentVersion: meta["version"],
		CurrentSource:  meta["source_name"],
		CurrentURL:     meta["source_url"],
		CurrentHash:    meta["sha256"],
		AutoCheck:      s.GetSettings().AutoCheck,
		AutoImport:     s.GetSettings().AutoImport,
	}
	if st.CurrentVersion == "" || st.CurrentSource == "" {
		if n, _ := store.CountQuestions(s.db); n > 0 {
			if st.CurrentVersion == "" {
				st.CurrentVersion = "本地题库"
			}
			if st.CurrentSource == "" {
				st.CurrentSource = "本地导入"
			}
		}
	}
	if ms, _ := strconv.ParseInt(meta["imported_at"], 10, 64); ms > 0 {
		st.CurrentImported = time.UnixMilli(ms)
	}
	return st
}

func (s *UpdateService) snapshotLocked() *models.BankUpdateStatus {
	st := s.statusFromMeta()
	if s.status != nil {
		st.Checked = s.status.Checked
		st.Checking = s.status.Checking
		st.Available = s.status.Available
		st.RemoteVersion = s.status.RemoteVersion
		st.RemoteURL = s.status.RemoteURL
		st.RemoteSize = s.status.RemoteSize
		st.RemoteModified = s.status.RemoteModified
		st.Message = s.status.Message
		st.Error = s.status.Error
	}
	return st
}

// GetStatus 返回缓存的检查结果与当前题库信息。
func (s *UpdateService) GetStatus() *models.BankUpdateStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshotLocked()
}

// ServiceStartup 启动后台自动检查。
func (s *UpdateService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	cleanupUpdateTemp()
	settings := s.GetSettings()
	s.mu.Lock()
	s.status = s.statusFromMeta()
	s.status.AutoCheck = settings.AutoCheck
	s.status.AutoImport = settings.AutoImport
	s.mu.Unlock()

	if settings.AutoCheck {
		go func() {
			select {
			case <-time.After(3 * time.Second):
			case <-ctx.Done():
				return
			}
			s.autoCheck()
		}()
	}
	return nil
}

// Check 抓取 CRAC 官方题库页面，比较远程 ZIP 与本地记录的版本。
func (s *UpdateService) Check() (*models.BankUpdateStatus, error) {
	s.mu.Lock()
	if s.checking {
		// 已有检查在进行：等待它完成，避免前端拿到“未检查完”的状态。
		wait := s.checkWait
		s.mu.Unlock()
		if wait != nil {
			<-wait
		}
		s.mu.Lock()
		st := s.snapshotLocked()
		s.mu.Unlock()
		return st, nil
	}
	s.checking = true
	s.checkWait = make(chan struct{})
	if s.status == nil {
		s.status = s.statusFromMeta()
	}
	s.status.Checking = true
	s.mu.Unlock()
	s.emitStatus()

	st, pending, err := s.doCheck()

	s.mu.Lock()
	s.checking = false
	if s.checkWait != nil {
		close(s.checkWait)
		s.checkWait = nil
	}
	if s.status == nil {
		s.status = s.statusFromMeta()
	}
	s.status.Checked = true
	s.status.Checking = false
	s.status.AutoCheck = s.GetSettings().AutoCheck
	s.status.AutoImport = s.GetSettings().AutoImport
	// 即使没有更新，也把远程版本信息带回状态，供界面展示。
	s.status.RemoteVersion = st.RemoteVersion
	s.status.RemoteURL = st.RemoteURL
	s.status.RemoteSize = st.RemoteSize
	s.status.RemoteModified = st.RemoteModified
	if err != nil {
		s.status.Available = false
		s.status.Error = err.Error()
		s.status.Message = "检查更新失败"
		if s.pendingPath != "" {
			_ = os.Remove(s.pendingPath)
			s.pendingPath = ""
			s.pending = remoteBank{}
		}
	} else {
		s.status.Error = ""
		s.status.Available = pending != nil
		s.status.Message = "已是最新题库"
		if pending != nil {
			s.status.Message = "发现新版本题库"
			s.status.RemoteVersion = pending.Version
			s.status.RemoteURL = pending.URL
			s.status.RemoteSize = pending.Size
			s.status.RemoteModified = pending.LastModified
			// 替换旧的待更新文件
			if s.pendingPath != "" && s.pendingPath != pending.TempPath {
				_ = os.Remove(s.pendingPath)
			}
			s.pendingPath = pending.TempPath
			s.pending = *pending
		} else {
			if s.pendingPath != "" {
				_ = os.Remove(s.pendingPath)
				s.pendingPath = ""
				s.pending = remoteBank{}
			}
		}
	}
	out := s.snapshotLocked()
	s.mu.Unlock()
	s.emitStatus()
	return out, err
}

func (s *UpdateService) doCheck() (*models.BankUpdateStatus, *remoteBank, error) {
	meta, _ := store.GetMetaMap(s.db, "bank.")
	st := s.statusFromMeta()

	pageURL, zipURL, version, err := discoverOfficialBank(meta["discovery_url"])
	if err != nil {
		return st, nil, err
	}

	remote := &remoteBank{URL: zipURL, PageURL: pageURL, Version: version}
	_ = store.SetMeta(s.db, store.MetaBankDiscoveryURL, pageURL)
	if err := headRemote(remote); err != nil {
		utils.Warn("UpdateService", "读取远程题库信息失败，将直接下载校验", "error", err.Error())
	}
	st.RemoteVersion = remote.Version
	st.RemoteURL = remote.URL
	st.RemoteSize = remote.Size
	st.RemoteModified = remote.LastModified

	// 题库为空时（首次启动/手动清空），无论历史哈希如何都应视为可更新。
	var questionCount int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM questions`).Scan(&questionCount)
	storedHash := meta["sha256"]
	storedURL := meta["source_url"]
	storedSize := meta["size"]
	storedETag := meta["etag"]
	storedLM := meta["last_modified"]
	if questionCount == 0 {
		storedHash = ""
		storedURL = ""
		storedSize = ""
		storedETag = ""
		storedLM = ""
	}

	sizeMatches := remote.Size <= 0 || storedSize == strconv.FormatInt(remote.Size, 10)
	if storedURL == remote.URL && storedHash != "" && sizeMatches {
		if remote.ETag != "" && storedETag == remote.ETag {
			_ = store.SetMetas(s.db, map[string]string{
				store.MetaBankVersion:      remote.Version,
				store.MetaBankDiscoveryURL: remote.PageURL,
			})
			st.Message = "已是最新题库"
			return st, nil, nil
		}
		if remote.LastModified != "" && storedLM == remote.LastModified {
			_ = store.SetMetas(s.db, map[string]string{
				store.MetaBankVersion:      remote.Version,
				store.MetaBankDiscoveryURL: remote.PageURL,
			})
			st.Message = "已是最新题库"
			return st, nil, nil
		}
	}

	path, hash, size, err := downloadZip(remote.URL)
	if err != nil {
		return st, nil, err
	}
	remote.Hash = hash
	remote.Size = size
	remote.TempPath = path
	st.RemoteSize = size
	st.RemoteModified = remote.LastModified

	if storedHash != "" && hash == storedHash {
		// 内容未变，更新来源元数据后视为最新
		_ = store.SetMetas(s.db, map[string]string{
			store.MetaBankSourceURL:    remote.URL,
			store.MetaBankSourceName:   filepath.Base(remote.URL),
			store.MetaBankETag:         remote.ETag,
			store.MetaBankLastModified: remote.LastModified,
			store.MetaBankSize:         strconv.FormatInt(size, 10),
			store.MetaBankVersion:      remote.Version,
			store.MetaBankDiscoveryURL: remote.PageURL,
		})
		_ = os.Remove(path)
		st.Message = "已是最新题库"
		return st, nil, nil
	}

	st.Message = "发现新版本题库"
	return st, remote, nil
}

func (s *UpdateService) autoCheck() {
	st, err := s.Check()
	if err != nil || st == nil || !st.Available || !st.AutoImport {
		return
	}
	// 有进行中的考试时不做自动替换，避免破坏正在作答的试卷。
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM exam_records WHERE status = 0`).Scan(&n); err != nil {
		return
	}
	if n > 0 {
		s.mu.Lock()
		if s.status != nil {
			s.status.Message = "发现新题库，但有进行中的考试，已延后自动更新"
		}
		s.mu.Unlock()
		s.emitStatus()
		return
	}
	if _, err := s.Apply(); err != nil {
		utils.Error("UpdateService", "自动更新题库失败", err)
	}
}

// Apply 下载（如有必要）并导入最新题库。用户数据按 J/P/I 尽量保留。
func (s *UpdateService) Apply() (*models.ImportResult, error) {
	s.mu.Lock()
	path := s.pendingPath
	remote := s.pending
	s.mu.Unlock()

	if path == "" {
		st, err := s.Check()
		if err != nil {
			return nil, err
		}
		if st == nil || !st.Available {
			return nil, fmt.Errorf("当前已是最新题库")
		}
		s.mu.Lock()
		path = s.pendingPath
		remote = s.pending
		s.mu.Unlock()
	}
	if path == "" {
		return nil, fmt.Errorf("更新文件不可用，请重新检查")
	}

	result, err := NewImportService(s.db).ImportZip(path)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	values := map[string]string{
		store.MetaBankSourceURL:    remote.URL,
		store.MetaBankSourceName:   filepath.Base(remote.URL),
		store.MetaBankETag:         remote.ETag,
		store.MetaBankLastModified: remote.LastModified,
		store.MetaBankSize:         strconv.FormatInt(remote.Size, 10),
		store.MetaBankVersion:      remote.Version,
		store.MetaBankDiscoveryURL: remote.PageURL,
		store.MetaBankImportedAt:   strconv.FormatInt(now.UnixMilli(), 10),
		store.MetaBankSHA256:       remote.Hash,
	}
	if err := store.SetMetas(s.db, values); err != nil {
		utils.Warn("UpdateService", "写入题库来源信息失败", "error", err.Error())
	}
	_ = os.Remove(path)

	s.mu.Lock()
	s.pendingPath = ""
	s.pending = remoteBank{}
	s.status = s.statusFromMeta()
	s.status.Checked = true
	s.status.Available = false
	s.status.Message = "题库已更新到最新版本"
	s.status.RemoteVersion = remote.Version
	s.status.RemoteURL = remote.URL
	s.status.RemoteSize = remote.Size
	s.status.RemoteModified = remote.LastModified
	s.mu.Unlock()
	s.emitStatus()

	utils.Info("UpdateService", "题库更新完成", "version", remote.Version, "questions", result.Imported)
	return result, nil
}

// cleanupUpdateTemp 清理上次运行遗留的待更新 ZIP。
func cleanupUpdateTemp() {
	files, err := filepath.Glob(filepath.Join(config.ImportTempDir(), "bank-update-*.zip"))
	if err != nil {
		return
	}
	for _, f := range files {
		_ = os.Remove(f)
	}
}

func (s *UpdateService) emitStatus() {
	app := application.Get()
	if app == nil {
		return
	}
	s.mu.Lock()
	st := s.snapshotLocked()
	s.mu.Unlock()
	app.Event.Emit("bank:update-status", st)
}

const officialCategoryPage = "https://www.crac.org.cn/?cat=7"

var (
	anchorRe  = regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	titleRe   = regexp.MustCompile(`(?is)<h1[^>]*class=["'][^"']*post-title[^"']*["'][^>]*>(.*?)</h1>`)
	pageTagRe = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
	zipDateRe = regexp.MustCompile(`(20\d{6})\d{0,6}_`)
)

type bankCandidate struct {
	zipURL  string
	pageURL string
	title   string
	date    time.Time
	dateStr string
}

// discoverOfficialBank 不只盯一个固定页面：
// 同时检查已知题库页、资料下载栏目、站内搜索，并跟随其中标题相关的文章，
// 取“日期最新且带 ZIP 附件”的候选。这样官方把新题库换到新页面也能发现。
func discoverOfficialBank(storedPage string) (pageURL, zipURL, version string, err error) {
	sources := []string{}
	if strings.TrimSpace(storedPage) != "" {
		sources = append(sources, storedPage)
	}
	sources = append(sources,
		officialBankPage,
		officialCategoryPage,
		"https://www.crac.org.cn/",
		"https://www.crac.org.cn/?s="+url.QueryEscape("题库"),
	)
	seenSource := map[string]bool{}
	postPages := map[string]bool{}
	var candidates []bankCandidate

	addCandidate := func(page, body string) {
		zip := chooseZipLink(body, page)
		if zip == "" {
			return
		}
		title := extractPageTitle(body)
		date, dateStr := extractPageDate(body)
		if zd := dateFromZipURL(zip); zd.After(date) {
			date = zd
			dateStr = zd.Format("2006-01-02")
		}
		candidates = append(candidates, bankCandidate{
			zipURL: zip, pageURL: page, title: title, date: date, dateStr: dateStr,
		})
	}

	// 题库页、资料下载栏目、站内搜索并发抓取，显著降低等待时间。
	bodies := fetchPages(sources, 4)
	for _, src := range sources {
		if seenSource[src] {
			continue
		}
		seenSource[src] = true
		body, ok := bodies[src]
		if !ok {
			continue
		}
		// 已知题库页本身通常直接带 ZIP
		addCandidate(src, body)
		// 资料下载/搜索结果里的相关文章
		for _, p := range extractPostLinks(body, src) {
			if len(postPages) >= 10 {
				break
			}
			postPages[p] = true
		}
	}

	// 并发抓取相关文章页（最多 10 个）。
	var pages []string
	for p := range postPages {
		if seenSource[p] {
			continue
		}
		pages = append(pages, p)
		if len(pages) >= 10 {
			break
		}
	}
	postBodies := fetchPages(pages, 4)
	for _, p := range pages {
		seenSource[p] = true
		if body, ok := postBodies[p]; ok {
			addCandidate(p, body)
		}
	}

	if len(candidates) == 0 {
		return "", "", "", fmt.Errorf("未在 CRAC 官网找到题库 ZIP（已检查题库页、资料下载与站内搜索）")
	}

	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.date.After(best.date) {
			best = c
		} else if c.date.Equal(best.date) &&
			strings.Contains(c.title, "题库") && !strings.Contains(best.title, "题库") {
			best = c
		}
	}

	version = best.title
	if version == "" {
		version = filepath.Base(best.zipURL)
	}
	if best.dateStr != "" {
		version += " · " + best.dateStr
	}
	if r := []rune(version); len(r) > 120 {
		version = string(r[:120]) + "…"
	}
	return best.pageURL, best.zipURL, version, nil
}

var cracClient = &http.Client{Timeout: 20 * time.Second}

// fetchPages 并发抓取一组页面，失败的页面直接跳过。
func fetchPages(urls []string, workers int) map[string]string {
	out := map[string]string{}
	if len(urls) == 0 {
		return out
	}
	if workers < 1 {
		workers = 1
	}
	type result struct {
		url  string
		body string
		err  error
	}
	results := make(chan result, len(urls))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			body, err := fetchPage(u)
			results <- result{url: u, body: body, err: err}
		}(u)
	}
	wg.Wait()
	close(results)
	for r := range results {
		if r.err == nil {
			out[r.url] = r.body
		}
	}
	return out
}

func fetchPage(rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "crac-exam-v3/"+config.AppVersion)
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	resp, err := cracClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("访问 %s 失败：%w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func extractPostLinks(body, baseURL string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range anchorRe.FindAllStringSubmatch(body, -1) {
		href := strings.TrimSpace(m[1])
		if !strings.Contains(href, "?p=") {
			continue
		}
		title := stripHTML(m[2])
		if !isBankTitle(title) {
			continue
		}
		resolved := resolveCRACURL(baseURL, href)
		if resolved == "" || seen[resolved] {
			continue
		}
		seen[resolved] = true
		out = append(out, resolved)
	}
	return out
}

func isBankTitle(title string) bool {
	t := strings.TrimSpace(title)
	if t == "" {
		return false
	}
	if strings.Contains(t, "题库") {
		return true
	}
	if strings.Contains(t, "操作技术能力") && strings.Contains(t, "验证") &&
		!strings.Contains(t, "考核") && !strings.Contains(t, "报名") {
		return true
	}
	if strings.Contains(t, "考试程序") {
		return true
	}
	return false
}

func chooseZipLink(body, baseURL string) string {
	var fallback string
	for _, m := range zipLinkRe.FindAllStringSubmatch(body, -1) {
		resolved := resolveCRACURL(baseURL, strings.TrimSpace(m[1]))
		if resolved == "" {
			continue
		}
		if fallback == "" {
			fallback = resolved
		}
		text := stripHTML(m[2])
		if strings.Contains(text, "附件1") || strings.Contains(text, "题库") ||
			strings.Contains(text, "2025") || strings.Contains(text, "2026") {
			return resolved
		}
	}
	return fallback
}

func stripHTML(s string) string {
	s = regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

func extractPageTitle(body string) string {
	if m := titleRe.FindStringSubmatch(body); len(m) == 2 {
		return stripHTML(m[1])
	}
	if m := pageTagRe.FindStringSubmatch(body); len(m) == 2 {
		t := stripHTML(m[1])
		t = strings.TrimSuffix(t, " | CRAC")
		t = strings.TrimSuffix(t, " – CRAC")
		return t
	}
	return ""
}

func extractPageDate(body string) (time.Time, string) {
	m := postDateRe.FindStringSubmatch(body)
	if len(m) != 2 {
		return time.Time{}, ""
	}
	s := stripHTML(m[1])
	for _, layout := range []string{"2006年1月2日", "2006年01月02日", "2006-01-02", "2006/01/02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, t.Format("2006-01-02")
		}
	}
	return time.Time{}, s
}

func dateFromZipURL(rawURL string) time.Time {
	m := zipDateRe.FindStringSubmatch(rawURL)
	if len(m) != 2 {
		return time.Time{}
	}
	if t, err := time.ParseInLocation("20060102", m[1], time.Local); err == nil {
		return t
	}
	return time.Time{}
}

func resolveCRACURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	host := strings.ToLower(abs.Hostname())
	if host != "crac.org.cn" && host != "www.crac.org.cn" {
		return ""
	}
	if abs.Scheme != "https" && abs.Scheme != "http" {
		return ""
	}
	return abs.String()
}

func headRemote(remote *remoteBank) error {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodHead, remote.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "crac-exam-v3/"+config.AppVersion)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	remote.Size = resp.ContentLength
	if remote.Size < 0 {
		remote.Size = 0
	}
	remote.ETag = resp.Header.Get("ETag")
	remote.LastModified = resp.Header.Get("Last-Modified")
	return nil
}

func downloadZip(rawURL string) (path, hash string, size int64, err error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", "", 0, err
	}
	req.Header.Set("User-Agent", "crac-exam-v3/"+config.AppVersion)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("下载题库失败：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("下载题库失败：HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > 600<<20 {
		return "", "", 0, fmt.Errorf("题库文件过大")
	}
	f, err := os.CreateTemp(config.ImportTempDir(), "bank-update-*.zip")
	if err != nil {
		return "", "", 0, err
	}
	path = f.Name()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, 600<<20+1))
	closeErr := f.Close()
	if err != nil {
		_ = os.Remove(path)
		return "", "", 0, err
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return "", "", 0, closeErr
	}
	if n > 600<<20 {
		_ = os.Remove(path)
		return "", "", 0, fmt.Errorf("题库文件过大")
	}
	return path, hex.EncodeToString(h.Sum(nil)), n, nil
}
