package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	githubAPIBase = "https://api.github.com"
	giteeAPIBase  = "https://gitee.com/api/v5"
	maxUpdateSize = int64(1 << 30) // 1 GiB，防止异常 Release 附件占满磁盘

	// UpdateSourceGitHub / UpdateSourceGitee 是应用更新可选的两个下载通道。
	UpdateSourceGitHub = "github"
	UpdateSourceGitee  = "gitee"
)

var (
	githubRepoRe    = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	updateNameRe    = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
	githubHostNames = []string{"github.com", "api.github.com"}
	giteeHostNames  = []string{"gitee.com", "www.gitee.com"}
	updateUserAgent = "crac-exam-v3/" + config.AppVersion
)

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
	Digest             string `json:"digest"`
}

type githubRelease struct {
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Body        string               `json:"body"`
	HTMLURL     string               `json:"html_url"`
	PublishedAt time.Time            `json:"published_at"`
	Assets      []githubReleaseAsset `json:"assets"`
}

// giteeRelease 是 Gitee OpenAPI 的 Release 结构，字段比 GitHub 少。
type giteeRelease struct {
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Body        string               `json:"body"`
	Prerelease  bool                 `json:"prerelease"`
	CreatedAt   string               `json:"created_at"`
	Assets      []githubReleaseAsset `json:"assets"`
}

// AppUpdateService 通过 GitHub 或 Gitee Release 检查、下载并启动应用更新。
type AppUpdateService struct {
	db          *sql.DB
	mu          sync.Mutex
	status      models.AppUpdateStatus
	checking    bool
	downloading bool
	client      *http.Client
	apiBase     string // GitHub API 基址（测试可覆盖）
	giteeBase   string // Gitee API 基址（测试可覆盖）
}

func NewAppUpdateService(db *sql.DB) *AppUpdateService {
	return &AppUpdateService{
		db:        db,
		client:    newUpdateHTTPClient(30 * time.Second),
		apiBase:   githubAPIBase,
		giteeBase: giteeAPIBase,
		status: models.AppUpdateStatus{
			CurrentVersion: config.AppVersion,
			Source:         UpdateSourceGitHub,
			Repo:           config.DefaultGitHubRepo,
			Assets:         []models.AppReleaseAsset{},
			Message:        "尚未检查更新",
		},
	}
}

func (s *AppUpdateService) ServiceName() string { return "AppUpdateService" }

func (s *AppUpdateService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.mu.Lock()
	s.status.CurrentVersion = config.AppVersion
	s.status.Source = s.source()
	s.status.Repo = s.repo()
	if s.status.Assets == nil {
		s.status.Assets = []models.AppReleaseAsset{}
	}
	s.mu.Unlock()
	return nil
}

// GetStatus 返回当前应用更新状态。
func (s *AppUpdateService) GetStatus() *models.AppUpdateStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.snapshotLocked()
	return &st
}

// SetRepo 修改当前更新通道使用的仓库（owner/repo）。
func (s *AppUpdateService) SetRepo(repo string) error {
	repo = normalizeGitHubRepo(repo)
	if !githubRepoRe.MatchString(repo) {
		return errors.New("仓库格式应为 owner/repo")
	}
	if err := store.SetMeta(s.db, repoMetaKey(s.source()), repo); err != nil {
		return err
	}
	s.mu.Lock()
	s.status.Repo = repo
	s.status.Checked = false
	s.status.Available = false
	s.status.Error = ""
	s.status.Message = "更新源已保存"
	s.mu.Unlock()
	s.emitStatus()
	return nil
}

// SetSource 切换应用更新通道（GitHub / Gitee）。
func (s *AppUpdateService) SetSource(source string) error {
	src := normalizeUpdateSource(source)
	if err := store.SetMeta(s.db, store.MetaAppUpdateSource, src); err != nil {
		return err
	}
	s.mu.Lock()
	s.status.Source = src
	s.status.Repo = s.repo()
	s.status.Checked = false
	s.status.Checking = false
	s.status.Available = false
	s.status.LatestVersion = ""
	s.status.ReleaseName = ""
	s.status.ReleaseURL = ""
	s.status.PublishedAt = ""
	s.status.Notes = ""
	s.status.Asset = nil
	s.status.Assets = []models.AppReleaseAsset{}
	s.status.Error = ""
	s.status.Message = "更新通道已切换为 " + updateSourceLabel(src)
	s.mu.Unlock()
	s.emitStatus()
	return nil
}

// Check 请求 GitHub 最新 Release，并与当前版本比较。
func (s *AppUpdateService) Check() (*models.AppUpdateStatus, error) {
	s.mu.Lock()
	if s.checking {
		st := s.snapshotLocked()
		s.mu.Unlock()
		return &st, nil
	}
	s.checking = true
	s.status.Checking = true
	s.status.Error = ""
	s.status.Message = "正在检查更新…"
	s.mu.Unlock()
	s.emitStatus()

	release, err := s.fetchLatestRelease()

	s.mu.Lock()
	s.checking = false
	s.status.Checking = false
	s.status.Checked = true
	s.status.CurrentVersion = config.AppVersion
	s.status.Source = s.source()
	s.status.Repo = s.repo()
	if err != nil {
		s.status.Available = false
		s.status.LatestVersion = ""
		s.status.ReleaseName = ""
		s.status.ReleaseURL = ""
		s.status.PublishedAt = ""
		s.status.Notes = ""
		s.status.Assets = []models.AppReleaseAsset{}
		s.status.Asset = nil
		s.status.Error = err.Error()
		s.status.Message = "检查应用更新失败"
		st := s.snapshotLocked()
		s.mu.Unlock()
		s.emitStatus()
		return &st, err
	}

	assets := make([]models.AppReleaseAsset, 0, len(release.Assets))
	for _, a := range release.Assets {
		assets = append(assets, models.AppReleaseAsset{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
			Size:        a.Size,
			ContentType: a.ContentType,
			Digest:      a.Digest,
		})
	}
	asset := pickUpdateAsset(assets)
	available := compareVersions(release.TagName, config.AppVersion) > 0

	s.status.LatestVersion = release.TagName
	s.status.ReleaseName = release.Name
	s.status.ReleaseURL = release.HTMLURL
	s.status.PublishedAt = formatReleaseTime(release.PublishedAt)
	s.status.Notes = truncateText(release.Body, 6000)
	s.status.Assets = assets
	s.status.Asset = asset
	s.status.Available = available
	s.status.Error = ""
	if available {
		s.status.Message = "发现新版本 " + release.TagName
		if asset == nil {
			s.status.Message = "发现新版本，但 Release 中没有 Windows 安装包"
		}
	} else {
		s.status.Message = "已是最新版本（" + updateSourceLabel(s.status.Source) + " 通道）"
	}
	st := s.snapshotLocked()
	s.mu.Unlock()
	s.emitStatus()
	return &st, nil
}

// Download 下载当前 Release 中推荐的应用更新包。
func (s *AppUpdateService) Download() (*models.AppUpdateStatus, error) {
	s.mu.Lock()
	if s.downloading {
		st := s.snapshotLocked()
		s.mu.Unlock()
		return &st, nil
	}
	if s.status.Asset == nil || !s.status.Available {
		st := s.snapshotLocked()
		s.mu.Unlock()
		return &st, errors.New("没有可下载的更新包，请先检查更新或打开发布页")
	}
	asset := *s.status.Asset
	s.downloading = true
	s.status.Downloading = true
	s.status.DownloadPercent = 0
	s.status.Error = ""
	s.status.Message = "正在下载更新包…"
	s.mu.Unlock()
	s.emitStatus()

	path, err := s.downloadAsset(asset)

	s.mu.Lock()
	s.downloading = false
	s.status.Downloading = false
	if err != nil {
		s.status.Error = err.Error()
		s.status.Message = "下载更新包失败"
	} else {
		s.status.DownloadedPath = path
		s.status.DownloadedName = filepath.Base(path)
		s.status.DownloadPercent = 100
		s.status.Error = ""
		s.status.Message = "更新包下载完成"
	}
	st := s.snapshotLocked()
	s.mu.Unlock()
	s.emitStatus()
	if err == nil {
		s.emitProgress("done", "更新包下载完成", 100)
	}
	return &st, err
}

// RunInstaller 启动已经下载的安装程序。必须由用户显式点击触发。
func (s *AppUpdateService) RunInstaller() error {
	s.mu.Lock()
	path := s.status.DownloadedPath
	s.mu.Unlock()
	if path == "" {
		return errors.New("请先下载更新包")
	}
	if err := validateDownloadedPath(path); err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch strings.ToLower(filepath.Ext(path)) {
	case ".msi":
		cmd = exec.Command("msiexec", "/i", path)
	case ".exe":
		cmd = exec.Command(path)
	default:
		return errors.New("只支持运行 .exe 或 .msi 更新包")
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动安装程序失败：%w", err)
	}
	s.mu.Lock()
	s.status.Message = "安装程序已启动"
	s.mu.Unlock()
	s.emitStatus()
	return nil
}

// OpenReleasePage 用系统浏览器打开发布页。
func (s *AppUpdateService) OpenReleasePage() error {
	s.mu.Lock()
	raw := s.status.ReleaseURL
	s.mu.Unlock()
	if raw == "" {
		return errors.New("暂无 Release 页面")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || !isTrustedUpdateHost(u.Hostname()) {
		return errors.New("发布页地址无效")
	}
	return openExternalURL(raw)
}

// OpenDownloadFolder 在文件管理器中定位已下载的更新包。
func (s *AppUpdateService) OpenDownloadFolder() error {
	s.mu.Lock()
	path := s.status.DownloadedPath
	s.mu.Unlock()
	dir := config.UpdatesDir()
	if path != "" {
		dir = filepath.Dir(path)
	}
	return openLocalPath(dir)
}

// fetchLatestRelease 按当前更新通道抓取最新 Release。
func (s *AppUpdateService) fetchLatestRelease() (*githubRelease, error) {
	if s.source() == UpdateSourceGitee {
		return s.fetchGiteeRelease()
	}
	return s.fetchGitHubRelease()
}

func (s *AppUpdateService) fetchGitHubRelease() (*githubRelease, error) {
	repo := s.repo()
	if !githubRepoRe.MatchString(repo) {
		return nil, errors.New("GitHub 仓库格式无效，请填写 owner/repo")
	}
	endpoint := strings.TrimRight(s.apiBase, "/") + "/repos/" + repo + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", updateUserAgent)
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("访问 GitHub 失败：%w", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if readErr != nil {
		return nil, fmt.Errorf("读取 GitHub 响应失败：%w", readErr)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		// 继续解析
	case http.StatusNotFound:
		return nil, fmt.Errorf("未找到 Release：请确认 %s 已发布 Release", repo)
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, errors.New("GitHub API 请求受限，请稍后重试")
	default:
		return nil, fmt.Errorf("GitHub 返回 HTTP %d：%s", resp.StatusCode, truncateText(string(body), 200))
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("解析 GitHub Release 失败：%w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return nil, errors.New("最新 Release 没有版本号（tag_name）")
	}
	return &release, nil
}

// fetchGiteeRelease 抓取 Gitee 镜像仓库的最新 Release。
func (s *AppUpdateService) fetchGiteeRelease() (*githubRelease, error) {
	repo := s.repo()
	if !githubRepoRe.MatchString(repo) {
		return nil, errors.New("Gitee 仓库格式无效，请填写 owner/repo")
	}
	base := s.giteeBase
	if base == "" {
		base = giteeAPIBase
	}
	endpoint := strings.TrimRight(base, "/") + "/repos/" + repo + "/releases/latest"
	if token := strings.TrimSpace(os.Getenv("GITEE_TOKEN")); token != "" {
		endpoint += "?access_token=" + url.QueryEscape(token)
	}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", updateUserAgent)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("访问 Gitee 失败：%w", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if readErr != nil {
		return nil, fmt.Errorf("读取 Gitee 响应失败：%w", readErr)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		// 继续解析
	case http.StatusNotFound:
		return nil, fmt.Errorf("未找到 Release：请确认 %s 已在 Gitee 发布 Release 并上传安装包", repo)
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, errors.New("Gitee 请求受限，请稍后重试")
	default:
		return nil, fmt.Errorf("Gitee 返回 HTTP %d：%s", resp.StatusCode, truncateText(string(body), 200))
	}

	var rel giteeRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("解析 Gitee Release 失败：%w", err)
	}
	if strings.TrimSpace(rel.TagName) == "" {
		return nil, errors.New("最新 Release 没有版本号（tag_name）")
	}
	published := time.Time{}
	if v := strings.TrimSpace(rel.CreatedAt); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			published = t
		}
	}
	return &githubRelease{
		TagName:     rel.TagName,
		Name:        rel.Name,
		Body:        rel.Body,
		HTMLURL:     "https://gitee.com/" + repo + "/releases/tag/" + url.PathEscape(rel.TagName),
		PublishedAt: published,
		Assets:      rel.Assets,
	}, nil
}

func (s *AppUpdateService) downloadAsset(asset models.AppReleaseAsset) (string, error) {
	u, err := url.Parse(asset.DownloadURL)
	if err != nil || u.Scheme != "https" || !isTrustedUpdateHost(u.Hostname()) {
		return "", errors.New("更新包下载地址无效")
	}
	client := newUpdateHTTPClient(15 * time.Minute)
	req, err := http.NewRequest(http.MethodGet, asset.DownloadURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", updateUserAgent)
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载更新包失败：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载更新包失败：HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxUpdateSize {
		return "", errors.New("更新包超过 1 GiB，已停止下载")
	}

	dir := config.UpdatesDir()
	name := safeUpdateName(asset.Name)
	path := filepath.Join(dir, name)
	part := path + ".part"
	f, err := os.Create(part)
	if err != nil {
		return "", fmt.Errorf("创建更新包文件失败：%w", err)
	}

	total := resp.ContentLength
	hash := sha256.New()
	writer := &updateProgressWriter{
		writer: io.MultiWriter(f, hash),
		total:  total,
		last:   time.Now(),
		onProgress: func(written, total int64) {
			percent := 0
			if total > 0 {
				percent = int(written * 100 / total)
				if percent > 99 {
					percent = 99
				}
			}
			s.setDownloadPercent(percent)
		},
	}
	n, copyErr := io.Copy(writer, io.LimitReader(resp.Body, maxUpdateSize+1))
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(part)
		return "", fmt.Errorf("下载更新包失败：%w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(part)
		return "", fmt.Errorf("保存更新包失败：%w", closeErr)
	}
	if n > maxUpdateSize {
		_ = os.Remove(part)
		return "", errors.New("更新包超过 1 GiB，已停止下载")
	}
	if digest := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(asset.Digest), "sha256:")); digest != "" {
		if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), digest) {
			_ = os.Remove(part)
			return "", errors.New("更新包 SHA-256 校验失败")
		}
	}
	_ = os.Remove(path)
	if err := os.Rename(part, path); err != nil {
		_ = os.Remove(part)
		return "", fmt.Errorf("保存更新包失败：%w", err)
	}
	return path, nil
}

// source 返回当前更新通道（github / gitee）。
func (s *AppUpdateService) source() string {
	if v, err := store.GetMeta(s.db, store.MetaAppUpdateSource); err == nil && strings.TrimSpace(v) != "" {
		return normalizeUpdateSource(v)
	}
	return UpdateSourceGitHub
}

// repo 返回当前更新通道对应的仓库（owner/repo）。
func (s *AppUpdateService) repo() string {
	key, env, def := repoMetaKey(s.source()), "CRAC_EXAM_GITHUB_REPO", config.DefaultGitHubRepo
	if s.source() == UpdateSourceGitee {
		env, def = "CRAC_EXAM_GITEE_REPO", config.DefaultGiteeRepo
	}
	if v, err := store.GetMeta(s.db, key); err == nil {
		if normalized := normalizeGitHubRepo(v); normalized != "" {
			return normalized
		}
	}
	if v := strings.TrimSpace(os.Getenv(env)); v != "" {
		return normalizeGitHubRepo(v)
	}
	return def
}

func repoMetaKey(source string) string {
	if normalizeUpdateSource(source) == UpdateSourceGitee {
		return store.MetaAppGiteeRepo
	}
	return store.MetaAppGitHubRepo
}

func normalizeUpdateSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "gitee", "gitee.com", "码云", "cn":
		return UpdateSourceGitee
	default:
		return UpdateSourceGitHub
	}
}

func updateSourceLabel(source string) string {
	if normalizeUpdateSource(source) == UpdateSourceGitee {
		return "Gitee 镜像"
	}
	return "GitHub"
}

func (s *AppUpdateService) snapshotLocked() models.AppUpdateStatus {
	st := s.status
	if st.CurrentVersion == "" {
		st.CurrentVersion = config.AppVersion
	}
	if st.Repo == "" {
		st.Repo = s.repo()
	}
	if st.Assets == nil {
		st.Assets = []models.AppReleaseAsset{}
	}
	return st
}

func (s *AppUpdateService) emitStatus() {
	s.mu.Lock()
	st := s.snapshotLocked()
	s.mu.Unlock()
	if app := application.Get(); app != nil {
		app.Event.Emit("app:update-status", st)
	}
}

func (s *AppUpdateService) emitProgress(stage, message string, percent int) {
	if app := application.Get(); app != nil {
		app.Event.Emit("app:update-progress", models.AppUpdateProgress{
			Stage:   stage,
			Message: message,
			Percent: percent,
		})
	}
}

func (s *AppUpdateService) setDownloadPercent(percent int) {
	s.mu.Lock()
	s.status.DownloadPercent = percent
	s.mu.Unlock()
	s.emitProgress("download", "正在下载更新包…", percent)
}

type updateProgressWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	last       time.Time
	onProgress func(written, total int64)
}

func (w *updateProgressWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.written += int64(n)
	if w.onProgress != nil && time.Since(w.last) >= 150*time.Millisecond {
		w.last = time.Now()
		w.onProgress(w.written, w.total)
	}
	return n, err
}

func newUpdateHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("更新下载重定向次数过多")
			}
			if !isTrustedUpdateHost(req.URL.Hostname()) {
				return fmt.Errorf("更新下载被重定向到不受信任的地址：%s", req.URL.Hostname())
			}
			return nil
		},
	}
}

func isGitHubHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	for _, allowed := range githubHostNames {
		if host == allowed {
			return true
		}
	}
	return strings.HasSuffix(host, ".githubusercontent.com")
}

// isTrustedUpdateHost 同时信任 GitHub 与 Gitee（更新通道二选一）。
func isTrustedUpdateHost(host string) bool {
	if isGitHubHost(host) {
		return true
	}
	host = strings.ToLower(strings.TrimSpace(host))
	for _, allowed := range giteeHostNames {
		if host == allowed {
			return true
		}
	}
	return strings.HasSuffix(host, ".gitee.com")
}

func normalizeGitHubRepo(repo string) string {
	repo = strings.TrimSpace(repo)
	for _, prefix := range []string{
		"https://github.com/", "http://github.com/", "github.com/",
		"https://gitee.com/", "http://gitee.com/", "gitee.com/",
	} {
		repo = strings.TrimPrefix(repo, prefix)
	}
	repo = strings.TrimSuffix(repo, ".git")
	return strings.Trim(repo, "/")
}

func pickUpdateAsset(assets []models.AppReleaseAsset) *models.AppReleaseAsset {
	var best *models.AppReleaseAsset
	bestScore := -1
	for i := range assets {
		a := assets[i]
		name := strings.ToLower(strings.TrimSpace(a.Name))
		if a.DownloadURL == "" || name == "" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".exe" && ext != ".msi" {
			continue
		}
		score := 0
		if strings.Contains(name, "windows") || strings.Contains(name, "win") {
			score += 12
		}
		if ext == ".exe" {
			score += 6
		} else {
			score += 4
		}
		if strings.Contains(name, "amd64") || strings.Contains(name, "x64") || strings.Contains(name, "x86_64") {
			score += 5
		}
		if strings.Contains(name, "arm64") || strings.Contains(name, "aarch64") {
			score -= 10
		}
		if strings.Contains(name, "386") || strings.Contains(name, "x86") {
			score -= 3
		}
		// 同时提供便携版与安装器时，优先让用户运行安装器。
		if strings.Contains(name, "installer") || strings.Contains(name, "setup") {
			score += 3
		}
		if a.Size > 0 {
			score++
		}
		if score > bestScore {
			copy := a
			best = &copy
			bestScore = score
		}
	}
	return best
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	return v
}

func versionParts(v string) (nums []int, pre string, ok bool) {
	v = normalizeVersion(v)
	if v == "" {
		return nil, "", false
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		pre = v[i+1:]
		v = v[:i]
	}
	for _, p := range strings.Split(v, ".") {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, "", false
		}
		nums = append(nums, n)
	}
	return nums, pre, true
}

func compareVersions(a, b string) int {
	ap, apre, aok := versionParts(a)
	bp, bpre, bok := versionParts(b)
	if !aok || !bok {
		return strings.Compare(normalizeVersion(a), normalizeVersion(b))
	}
	max := len(ap)
	if len(bp) > max {
		max = len(bp)
	}
	for i := 0; i < max; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	if apre == bpre {
		return 0
	}
	if apre == "" {
		return 1
	}
	if bpre == "" {
		return -1
	}
	return strings.Compare(apre, bpre)
}

func safeUpdateName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = updateNameRe.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	if name == "" {
		return "crac-exam-update.exe"
	}
	return name
}

func validateDownloadedPath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return errors.New("更新包路径无效")
	}
	dir, err := filepath.Abs(config.UpdatesDir())
	if err != nil {
		return errors.New("更新目录无效")
	}
	rel, err := filepath.Rel(dir, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return errors.New("更新包路径无效")
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return errors.New("更新包不存在")
	}
	return nil
}

func formatReleaseTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func truncateText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func openExternalURL(raw string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", raw).Start()
	case "darwin":
		return exec.Command("open", raw).Start()
	default:
		return exec.Command("xdg-open", raw).Start()
	}
}

func openLocalPath(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
