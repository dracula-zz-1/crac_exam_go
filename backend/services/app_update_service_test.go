package services

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

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
