package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/moonshot-ai/kimi-config-server/githubpkg"
	"github.com/moonshot-ai/kimi-config-server/starlarkpkg"
)

type platformVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Latest  bool   `json:"latest"`
	Legacy  bool   `json:"legacy"`
	Draft   bool   `json:"draft"`
}

var versionFilePattern = regexp.MustCompile(`^v([0-9]+)\.star$`)

// Handler is the main entry point for Vercel Serverless Functions
func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(204)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Initialize GitHub client
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPO")
	configPath := os.Getenv("GITHUB_CONFIG_PATH")

	if token == "" || repo == "" {
		http.Error(w, `{"error": "GitHub configuration missing"}`, http.StatusInternalServerError)
		return
	}

	gh := githubpkg.NewClient(token, repo, configPath)

	// Ensure sample scripts exist (best effort)
	_ = gh.EnsureSampleScripts()

	// Route dispatch
	path := r.URL.Path

	switch {
	case path == "/api/platforms":
		handlePlatforms(w, r, gh)
	case strings.HasPrefix(path, "/api/scripts/") && strings.HasSuffix(path, "/versions"):
		handleVersions(w, r, gh)
	case strings.HasPrefix(path, "/api/scripts/") && strings.HasSuffix(path, "/drafts"):
		handleDrafts(w, r, gh)
	case strings.HasPrefix(path, "/api/scripts/") && strings.HasSuffix(path, "/history"):
		handleHistory(w, r, gh)
	case strings.HasPrefix(path, "/api/scripts/"):
		handleScripts(w, r, gh)
	case path == "/api/preview":
		handlePreview(w, r)
	case path == "/v1/config":
		handleConfig(w, r, gh)
	default:
		http.NotFound(w, r)
	}
}

func handlePlatforms(w http.ResponseWriter, r *http.Request, gh *githubpkg.Client) {
	if r.Method != "GET" {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	files, err := gh.ListStarFiles()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var platforms []string
	for _, f := range files {
		name := strings.TrimSuffix(f, ".star")
		platforms = append(platforms, name)
	}
	for _, platform := range []string{"ios", "android"} {
		versionFiles, err := gh.ListStarFilesInDir(platform)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		if len(versionFiles) > 0 && !contains(platforms, platform) {
			platforms = append(platforms, platform)
		}
	}
	sort.Strings(platforms)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"platforms": platforms,
	})
}

func handleScripts(w http.ResponseWriter, r *http.Request, gh *githubpkg.Client) {
	// Extract platform from path: /api/scripts/:platform
	prefix := "/api/scripts/"
	platform := strings.TrimPrefix(r.URL.Path, prefix)
	platform = strings.Split(platform, "/")[0] // Remove any trailing segments

	if platform == "" {
		http.Error(w, `{"error": "platform required"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		version := r.URL.Query().Get("version")
		file := scriptFileForRequest(gh, platform, version)
		content, _, err := gh.GetFile(file)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, `{"error": "script not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"platform": platform,
			"version":  version,
			"path":     file,
			"content":  content,
		})

	case "POST":
		// Check if this is a publish request
		if strings.HasSuffix(r.URL.Path, "/publish") {
			var pubReq struct {
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&pubReq); err != nil {
				pubReq.Message = fmt.Sprintf("Publish %s", platform)
			}
			file := scriptFileForRequest(gh, platform, r.URL.Query().Get("version"))
			if err := gh.PublishFile(file, pubReq.Message); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "published",
			})
			return
		}

		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		// Regular save
		file := scriptFileForWrite(platform, r.URL.Query().Get("version"))
		if err := gh.SaveFile(file, req.Content); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "saved",
			"path":    file,
		})

	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleVersions(w http.ResponseWriter, r *http.Request, gh *githubpkg.Client) {
	if r.Method != "GET" {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	platform := scriptPathPlatform(r.URL.Path, "/versions")
	if platform == "" {
		http.Error(w, `{"error": "platform required"}`, http.StatusBadRequest)
		return
	}

	versions, err := listPlatformVersions(gh, platform)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	latest := latestVersion(versions)
	next := "v1"
	if latest != nil {
		next = incrementVersion(latest.Version)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"platform":    platform,
		"versions":    versions,
		"latest":      latest,
		"nextVersion": next,
	})
}

func handleDrafts(w http.ResponseWriter, r *http.Request, gh *githubpkg.Client) {
	if r.Method != "POST" {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	platform := scriptPathPlatform(r.URL.Path, "/drafts")
	if platform == "" {
		http.Error(w, `{"error": "platform required"}`, http.StatusBadRequest)
		return
	}

	versions, err := listPlatformVersions(gh, platform)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	latest := latestVersion(versions)
	if latest == nil {
		http.Error(w, `{"error": "latest script not found"}`, http.StatusNotFound)
		return
	}

	next := incrementVersion(latest.Version)
	target := versionedScriptFile(platform, next)
	content, _, err := gh.GetFile(latest.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"platform":       platform,
		"version":        next,
		"path":           target,
		"content":        content,
		"baseVersion":    latest.Version,
		"basePath":       latest.Path,
		"baseContent":    content,
		"latestVersion":  latest.Version,
		"workingVersion": next,
	})
}

func handleHistory(w http.ResponseWriter, r *http.Request, gh *githubpkg.Client) {
	if r.Method != "GET" {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Extract platform from path: /api/scripts/:platform/history
	prefix := "/api/scripts/"
	suffix := "/history"
	path := r.URL.Path
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, suffix)
	platform := strings.Split(path, "/")[0]

	if platform == "" {
		http.Error(w, `{"error": "platform required"}`, http.StatusBadRequest)
		return
	}

	commits, err := gh.History(scriptFileForRequest(gh, platform, r.URL.Query().Get("version")), 20)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"platform": platform,
		"commits":  commits,
	})
}

func handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Script string `json:"script"`
		Ctx    struct {
			Platform string `json:"platform"`
			Version  string `json:"version"`
			Language string `json:"language"`
			Region   string `json:"region"`
		} `json:"ctx"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	config, err := starlarkpkg.Execute(req.Script, starlarkpkg.EvalContext{
		Platform: req.Ctx.Platform,
		Version:  req.Ctx.Version,
		Language: req.Ctx.Language,
		Region:   req.Ctx.Region,
	})
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"config": nil,
			"error":  err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"config": config,
	})
}

func handleConfig(w http.ResponseWriter, r *http.Request, gh *githubpkg.Client) {
	if r.Method != "GET" {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	platform := r.URL.Query().Get("platform")
	version := r.URL.Query().Get("version")
	language := r.URL.Query().Get("lang")
	region := r.URL.Query().Get("region")

	if platform == "" || version == "" || language == "" {
		http.Error(w, `{"error": "missing required parameters: platform, version, lang"}`, http.StatusBadRequest)
		return
	}

	content, _, err := gh.GetFile(versionedScriptFile(platform, version))
	if err != nil && strings.Contains(err.Error(), "not found") {
		content, _, err = gh.GetFile(platform + ".star")
	}
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	config, err := starlarkpkg.Execute(content, starlarkpkg.EvalContext{
		Platform: platform,
		Version:  version,
		Language: language,
		Region:   region,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(config)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func scriptPathPlatform(path, suffix string) string {
	path = strings.TrimPrefix(path, "/api/scripts/")
	path = strings.TrimSuffix(path, suffix)
	path = strings.Trim(path, "/")
	return strings.Split(path, "/")[0]
}

func listPlatformVersions(gh *githubpkg.Client, platform string) ([]platformVersion, error) {
	files, err := gh.ListStarFilesInDir(platform)
	if err != nil {
		return nil, err
	}

	versions := make([]platformVersion, 0, len(files)+1)
	for _, file := range files {
		match := versionFilePattern.FindStringSubmatch(file)
		if match == nil {
			continue
		}
		version := "v" + match[1]
		versions = append(versions, platformVersion{
			Version: version,
			Path:    platform + "/" + file,
		})
	}

	if _, _, err := gh.GetFile(platform + ".star"); err == nil && !hasVersion(versions, "v1") {
		versions = append(versions, platformVersion{
			Version: "v1",
			Path:    platform + ".star",
			Legacy:  true,
		})
	}

	sort.Slice(versions, func(i, j int) bool {
		return versionNumber(versions[i].Version) < versionNumber(versions[j].Version)
	})
	if len(versions) > 0 {
		versions[len(versions)-1].Latest = true
	}

	return versions, nil
}

func latestVersion(versions []platformVersion) *platformVersion {
	if len(versions) == 0 {
		return nil
	}
	latest := versions[len(versions)-1]
	return &latest
}

func hasVersion(versions []platformVersion, version string) bool {
	for _, item := range versions {
		if item.Version == version {
			return true
		}
	}
	return false
}

func incrementVersion(version string) string {
	return "v" + strconv.Itoa(versionNumber(version)+1)
}

func versionNumber(version string) int {
	version = strings.TrimPrefix(version, "v")
	n, _ := strconv.Atoi(version)
	return n
}

func versionedScriptFile(platform, version string) string {
	return platform + "/" + strings.TrimSuffix(version, ".star") + ".star"
}

func scriptFileForRequest(gh *githubpkg.Client, platform, version string) string {
	if version == "" {
		versions, err := listPlatformVersions(gh, platform)
		if err == nil {
			if latest := latestVersion(versions); latest != nil {
				return latest.Path
			}
		}
		return platform + ".star"
	}

	file := versionedScriptFile(platform, version)
	if _, _, err := gh.GetFile(file); err == nil {
		return file
	}
	return platform + ".star"
}

func scriptFileForWrite(platform, version string) string {
	if version == "" {
		return platform + ".star"
	}
	return versionedScriptFile(platform, version)
}
