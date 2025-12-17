package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GitHubRelease represents a GitHub release response
type GitHubRelease struct {
	TagName    string    `json:"tag_name"`
	Name       string    `json:"name"`
	Draft      bool      `json:"draft"`
	Prerelease bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL    string    `json:"html_url"`
	Body       string    `json:"body"`
}

// UpdateInfo contains version update information
type UpdateInfo struct {
	Available      bool      `json:"available"`
	CurrentVersion string    `json:"currentVersion"`
	LatestVersion  string    `json:"latestVersion"`
	ReleaseURL     string    `json:"releaseURL"`
	ReleaseNotes   string    `json:"releaseNotes"`
	PublishedAt    time.Time `json:"publishedAt"`
}

type UpdateService struct {
	ctx            context.Context
	currentVersion string
	repoOwner      string
	repoName       string
	lastCheck      time.Time
	lastResult     *UpdateInfo
	mu             sync.RWMutex
}

func NewUpdateService(currentVersion, repoOwner, repoName string) *UpdateService {
	return &UpdateService{
		currentVersion: currentVersion,
		repoOwner:      repoOwner,
		repoName:       repoName,
	}
}

func (s *UpdateService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// CheckForUpdates checks GitHub API for latest release
func (s *UpdateService) CheckForUpdates() (*UpdateInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cache check results for 5 minutes to avoid rate limiting
	if time.Since(s.lastCheck) < 5*time.Minute && s.lastResult != nil {
		return s.lastResult, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", s.repoOwner, s.repoName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Mac-Dev-Cleaner")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		info := &UpdateInfo{
			Available:      false,
			CurrentVersion: s.currentVersion,
			LatestVersion:  s.currentVersion,
		}
		s.lastCheck = time.Now()
		s.lastResult = info
		return info, nil
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(s.currentVersion, "v")

	isNewer := compareVersions(latestVersion, currentVersion)

	info := &UpdateInfo{
		Available:      isNewer,
		CurrentVersion: s.currentVersion,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
	}

	s.lastCheck = time.Now()
	s.lastResult = info

	return info, nil
}

// compareVersions compares two semantic versions (without 'v' prefix)
// Returns true if v1 > v2
func compareVersions(v1, v2 string) bool {
	// Simple semantic version comparison
	// Format: major.minor.patch
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 > n2 {
			return true
		}
		if n1 < n2 {
			return false
		}
	}

	return false
}

// ClearCache clears the cached result
func (s *UpdateService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastResult = nil
	s.lastCheck = time.Time{}
}
