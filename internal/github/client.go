package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client wraps GitHub API operations for the config repo
type Client struct {
	token      string
	owner      string
	repo       string
	configPath string
	httpClient *http.Client
}

// Commit represents a git commit from GitHub API
type Commit struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// FileContent represents GitHub content API response
type fileContent struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	SHA     string `json:"sha"`
	Content string `json:"content"`
	Encoding string `json:"encoding"`
	Type    string `json:"type"`
}

// CommitItem represents a single commit from GitHub API
type commitItem struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name  string    `json:"name"`
			Date  time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

// NewClient creates a new GitHub API client
func NewClient(token, repo, configPath string) *Client {
	parts := strings.Split(repo, "/")
	owner := parts[0]
	repoName := parts[1]
	if len(parts) != 2 {
		owner = ""
		repoName = repo
	}
	return &Client{
		token:      token,
		owner:      owner,
		repo:       repoName,
		configPath: configPath,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) apiURL(path string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s/%s", c.owner, c.repo, c.configPath, path)
}

func (c *Client) commitsURL(file string) string {
	if file != "" {
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?path=%s/%s&per_page=50", c.owner, c.repo, c.configPath, file)
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?per_page=50", c.owner, c.repo)
}

func (c *Client) dirURL() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", c.owner, c.repo, c.configPath)
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return c.httpClient.Do(req)
}

// GetFile reads a file from the GitHub repo
func (c *Client) GetFile(name string) (content string, sha string, err error) {
	url := c.apiURL(name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", "", fmt.Errorf("file not found")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("github api error: %s - %s", resp.Status, string(body))
	}

	var fc fileContent
	if err := json.NewDecoder(resp.Body).Decode(&fc); err != nil {
		return "", "", err
	}

	if fc.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(fc.Content)
		if err != nil {
			return "", "", err
		}
		return string(decoded), fc.SHA, nil
	}
	return fc.Content, fc.SHA, nil
}

// SaveFile creates or updates a file in the GitHub repo (working tree only, no commit)
func (c *Client) SaveFile(name string, content string) error {
	// In GitHub API, we can't save to working tree without committing
	// So we save to a temporary branch or just keep it in memory
	// For this implementation, we'll create a commit on a "workspace" branch
	sha := ""
	existing, existingSHA, err := c.GetFile(name)
	if err == nil && existing != "" {
		sha = existingSHA
	}

	message := fmt.Sprintf("Update %s (workspace save)", name)
	return c.commitFile(name, content, message, sha)
}

// PublishFile commits a file change
func (c *Client) PublishFile(name string, message string) error {
	// Read current content
	content, sha, err := c.GetFile(name)
	if err != nil {
		return err
	}

	if message == "" {
		message = fmt.Sprintf("Publish %s", name)
	}

	return c.commitFile(name, content, message, sha)
}

func (c *Client) commitFile(name, content, message, sha string) error {
	url := c.apiURL(name)

	payload := map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
	}
	if sha != "" {
		payload["sha"] = sha
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api error: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// History returns commit history for a file
func (c *Client) History(name string, n int) ([]Commit, error) {
	url := c.commitsURL(name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error: %s - %s", resp.Status, string(body))
	}

	var items []commitItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var commits []Commit
	for i, item := range items {
		if i >= n {
			break
		}
		commits = append(commits, Commit{
			Hash:      item.SHA[:7],
			Message:   item.Commit.Message,
			Author:    item.Commit.Author.Name,
			Timestamp: item.Commit.Author.Date,
		})
	}

	return commits, nil
}

// ListStarFiles returns all .star files in the config directory
func (c *Client) ListStarFiles() ([]string, error) {
	url := c.dirURL()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Directory doesn't exist yet, return empty
		return []string{}, nil
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error: %s - %s", resp.Status, string(body))
	}

	var items []fileContent
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var files []string
	for _, item := range items {
		if item.Type == "file" && strings.HasSuffix(item.Name, ".star") {
			files = append(files, item.Name)
		}
	}

	return files, nil
}

// EnsureSampleScripts creates sample scripts if they don't exist
func (c *Client) EnsureSampleScripts() error {
	iosScript := `def build_config(ctx):
    config = {
        "system": {
            "greeting": "Hello",
            "cacheLimitMB": 100,
            "webviewRedirectAllowSchemes": ["kimi", "https"],
        },
        "taskbar": {
            "items": [
                {"type": "DEEP_RESEARCH", "title": "Deep Research"},
                {"type": "PHOTO_SOLVER", "title": "Photo Solver"},
            ],
        },
        "model": {
            "defaultModel": "kimi-k2",
        },
    }

    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"
        config["taskbar"]["items"][0]["title"] = "深度研究"
        config["taskbar"]["items"][1]["title"] = "拍照解题"

    if ctx.language == "ja":
        config["system"]["greeting"] = "こんにちは"

    if ctx.region == "overseas":
        config["system"]["domains"] = {"apiDomain": "api.kimi.moonshot.cn"}

    if ctx.version >= "2.5.5":
        config["system"]["kimiplusPlaza"] = {"enabled": True}
        config["taskbar"]["items"].append({"type": "KIMI_PLUS", "title": "Kimi+"})

    if ctx.version >= "2.6.0":
        config["model"]["defaultModel"] = "kimi-k2.5"

    return config
`

	androidScript := `def build_config(ctx):
    config = {
        "system": {
            "greeting": "Hello",
        },
    }

    if ctx.language == "zh":
        config["system"]["greeting"] = "你好"

    return config
`

	// Check if ios.star exists
	_, _, err := c.GetFile("ios.star")
	if err != nil {
		// Create it
		if err := c.commitFile("ios.star", iosScript, "Initial ios.star", ""); err != nil {
			return err
		}
	}

	// Check if android.star exists
	_, _, err = c.GetFile("android.star")
	if err != nil {
		// Create it
		if err := c.commitFile("android.star", androidScript, "Initial android.star", ""); err != nil {
			return err
		}
	}

	return nil
}
