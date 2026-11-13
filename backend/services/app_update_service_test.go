package services

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"crac_exam_v3/backend/config"
	"crac_exam_v3/backend/models"
	"crac_exam_v3/backend/store"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v3.0.0", "3.0.0", 0},
		{"3.0.1", "3.0.0", 1},
		{"3.0.0", "3.0.1", -1},
		{"v3.1", "3.0.9", 1},
		{"3.0.0", "3.0.0-beta.1", 1},
		{"3.0.0-beta.1", "3.0.0", -1},
		{"3.0.0-beta.2", "3.0.0-beta.1", 1},
		{"3.0.0+build.7", "3.0.0", 0},
		{"not-a-version", "3.0.0", 1},
	}
	for _, tt := range tests {
		if got := compareVersions(tt.a, tt.b); got != tt.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestPickUpdateAsset(t *testing.T) {
	assets := []models.AppReleaseAsset{
		{Name: "crac-exam-v3.1.0-linux-amd64.tar.gz", DownloadURL: "https://github.com/x/y/releases/download/v3.1.0/linux.tar.gz"},
		{Name: "crac-exam-v3.1.0-windows-amd64.exe", DownloadURL: "https://github.com/x/y/releases/download/v3.1.0/win.exe", Size: 100},
		{Name: "crac-exam-v3.1.0-windows-arm64.exe", DownloadURL: "https://github.com/x/y/releases/download/v3.1.0/arm.exe", Size: 100},
		{Name: "crac-exam-v3.1.0-windows-amd64.exe.sha256", DownloadURL: "https://github.com/x/y/releases/download/v3.1.0/win.exe.sha256"},
	}
	got := pickUpdateAsset(assets)
	if got == nil {
		t.Fatal("pickUpdateAsset returned nil")
	}
	if got.Name != "crac-exam-v3.1.0-windows-amd64.exe" {
		t.Fatalf("pickUpdateAsset picked %q", got.Name)
	}
}

func TestPickUpdateAssetPrefersInstaller(t *testing.T) {
	assets := []models.AppReleaseAsset{
		{Name: "crac-exam-v3.1.0-windows-amd64.exe", DownloadURL: "https://github.com/x/y/releases/download/v3.1.0/portable.exe", Size: 100},
		{Name: "crac-exam-v3.1.0-windows-amd64-installer.exe", DownloadURL: "https://github.com/x/y/releases/download/v3.1.0/setup.exe", Size: 100},
	}
	got := pickUpdateAsset(assets)
	if got == nil {
		t.Fatal("pickUpdateAsset returned nil")
	}
	if got.Name != "crac-exam-v3.1.0-windows-amd64-installer.exe" {
		t.Fatalf("pickUpdateAsset picked %q, want installer", got.Name)
	}
}

func TestNormalizeGitHubRepo(t *testing.T) {
	tests := map[string]string{
		"https://github.com/dracula-zz-1/crac_exam_go.git": "dracula-zz-1/crac_exam_go",
		"github.com/dracula-zz-1/crac_exam_go/":            "dracula-zz-1/crac_exam_go",
		"  dracula-zz-1/crac_exam_go  ":                    "dracula-zz-1/crac_exam_go",
	}
	for input, want := range tests {
		if got := normalizeGitHubRepo(input); got != want {
			t.Errorf("normalizeGitHubRepo(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSafeUpdateName(t *testing.T) {
	if got := safeUpdateName("../../evil name.exe"); got != "evil_name.exe" {
		t.Fatalf("safeUpdateName() = %q", got)
	}
	if got := safeUpdateName(""); got != "crac-exam-update.exe" {
		t.Fatalf("safeUpdateName(empty) = %q", got)
	}
}

func TestCheckParsesReleaseAndPicksAsset(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "app-update.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/dracula-zz-1/crac_exam_go/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v3.1.0",
			"name": "v3.1.0",
			"body": "release notes",
			"html_url": "https://github.com/dracula-zz-1/crac_exam_go/releases/tag/v3.1.0",
			"published_at": "2026-11-12T08:00:00Z",
			"assets": [
				{"name": "crac-exam-v3.1.0-linux-amd64.tar.gz", "browser_download_url": "https://github.com/x/y/linux.tar.gz", "size": 10},
				{"name": "crac-exam-v3.1.0-windows-amd64.exe", "browser_download_url": "https://github.com/x/y/win.exe", "size": 20}
			]
		}`))
	}))
	defer server.Close()

	svc := NewAppUpdateService(db)
	svc.apiBase = server.URL
	st, err := svc.Check()
	if err != nil {
		t.Fatal(err)
	}
	if !st.Available {
		t.Fatalf("expected update available, got %+v", st)
	}
	if st.LatestVersion != "v3.1.0" || st.Asset == nil || st.Asset.Name != "crac-exam-v3.1.0-windows-amd64.exe" {
		t.Fatalf("unexpected status: %+v", st)
	}
}

func TestNormalizeUpdateSource(t *testing.T) {
	tests := map[string]string{
		"gitee":     UpdateSourceGitee,
		"Gitee":     UpdateSourceGitee,
		" GITEE ":   UpdateSourceGitee,
		"github":    UpdateSourceGitHub,
		"":          UpdateSourceGitHub,
		"something": UpdateSourceGitHub,
	}
	for input, want := range tests {
		if got := normalizeUpdateSource(input); got != want {
			t.Errorf("normalizeUpdateSource(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeRepoStripsBothHosts(t *testing.T) {
	tests := map[string]string{
		"https://gitee.com/dracula_zz/crac_exam_go.git": "dracula_zz/crac_exam_go",
		"gitee.com/dracula_zz/crac_exam_go":             "dracula_zz/crac_exam_go",
		"dracula_zz/crac_exam_go":                       "dracula_zz/crac_exam_go",
	}
	for input, want := range tests {
		if got := normalizeGitHubRepo(input); got != want {
			t.Errorf("normalizeGitHubRepo(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsTrustedUpdateHost(t *testing.T) {
	trusted := []string{"github.com", "api.github.com", "objects.githubusercontent.com", "gitee.com", "www.gitee.com", "raw.gitee.com"}
	for _, h := range trusted {
		if !isTrustedUpdateHost(h) {
			t.Errorf("isTrustedUpdateHost(%q) = false, want true", h)
		}
	}
	for _, h := range []string{"evil.com", "gitee.com.evil.com", "github.com.evil.com", ""} {
		if isTrustedUpdateHost(h) {
			t.Errorf("isTrustedUpdateHost(%q) = true, want false", h)
		}
	}
}

func TestSourceSwitchAndGiteeCheck(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "app-update-gitee.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/dracula_zz/crac_exam_go/releases/latest" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v3.2.0",
			"name": "v3.2.0 镜像版",
			"body": "gitee notes",
			"prerelease": false,
			"created_at": "2026-11-12T08:00:00+08:00",
			"assets": [
				{"name": "crac-exam-v3.2.0-windows-amd64.exe", "browser_download_url": "https://gitee.com/dracula_zz/crac_exam_go/attach_files/1/download/win.exe"}
			]
		}`))
	}))
	defer server.Close()

	svc := NewAppUpdateService(db)
	svc.giteeBase = server.URL

	if err := svc.SetSource("gitee"); err != nil {
		t.Fatal(err)
	}
	if got := svc.source(); got != UpdateSourceGitee {
		t.Fatalf("source() = %q, want gitee", got)
	}
	if got := svc.repo(); got != "dracula_zz/crac_exam_go" {
		t.Fatalf("repo() = %q", got)
	}

	st, err := svc.Check()
	if err != nil {
		t.Fatal(err)
	}
	if st.Source != UpdateSourceGitee {
		t.Fatalf("status source = %q", st.Source)
	}
	if !st.Available || st.LatestVersion != "v3.2.0" {
		t.Fatalf("unexpected status: %+v", st)
	}
	if st.Asset == nil || st.Asset.Name != "crac-exam-v3.2.0-windows-amd64.exe" {
		t.Fatalf("unexpected asset: %+v", st.Asset)
	}
	if st.ReleaseURL != "https://gitee.com/dracula_zz/crac_exam_go/releases/tag/v3.2.0" {
		t.Fatalf("unexpected release url: %q", st.ReleaseURL)
	}

	// 切回 GitHub 后状态应重置，且仓库回到 GitHub 默认值
	if err := svc.SetSource("github"); err != nil {
		t.Fatal(err)
	}
	st = svc.GetStatus()
	if st.Source != UpdateSourceGitHub || st.Available || st.LatestVersion != "" {
		t.Fatalf("source reset failed: %+v", st)
	}
	if st.Repo != config.DefaultGitHubRepo {
		t.Fatalf("repo after reset = %q", st.Repo)
	}
}
