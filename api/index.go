package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/moonshot-ai/kimi-config-server/githubpkg"
	"github.com/moonshot-ai/kimi-config-server/starlarkpkg"
)

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
		content, _, err := gh.GetFile(platform + ".star")
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
			"content":  content,
		})

	case "POST":
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
			return
		}

		// Check if this is a publish request
		if strings.HasSuffix(r.URL.Path, "/publish") {
			var pubReq struct {
				Message string `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&pubReq); err != nil {
				pubReq.Message = fmt.Sprintf("Publish %s", platform)
			}
			if err := gh.PublishFile(platform+".star", pubReq.Message); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "published",
			})
			return
		}

		// Regular save
		if err := gh.SaveFile(platform+".star", req.Content); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "saved",
		})

	default:
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
	}
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

	commits, err := gh.History(platform+".star", 20)
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

	content, _, err := gh.GetFile(platform + ".star")
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
