package models

// AppReleaseAsset GitHub Release 中的一个附件。
type AppReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"downloadUrl"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
	Digest      string `json:"digest"`
}

// AppUpdateStatus 应用在线更新状态。
type AppUpdateStatus struct {
	Checked         bool              `json:"checked"`
	Checking        bool              `json:"checking"`
	Downloading     bool              `json:"downloading"`
	Available       bool              `json:"available"`
	CurrentVersion  string            `json:"currentVersion"`
	LatestVersion   string            `json:"latestVersion"`
	ReleaseName     string            `json:"releaseName"`
	ReleaseURL      string            `json:"releaseUrl"`
	PublishedAt     string            `json:"publishedAt"`
	Notes           string            `json:"notes"`
	Repo            string            `json:"repo"`
	Asset           *AppReleaseAsset  `json:"asset"`
	Assets          []AppReleaseAsset `json:"assets"`
	DownloadPercent int               `json:"downloadPercent"`
	DownloadedPath  string            `json:"downloadedPath"`
	DownloadedName  string            `json:"downloadedName"`
	Message         string            `json:"message"`
	Error           string            `json:"error"`
}

// AppUpdateProgress 应用更新下载进度事件载荷。
type AppUpdateProgress struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
	Percent int    `json:"percent"`
}
