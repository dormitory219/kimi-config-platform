package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Repo wraps a git repository
type Repo struct {
	path string
	repo *git.Repository
}

// Commit represents a git commit
type Commit struct {
	Hash      string    `json:"hash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// Open opens or initializes a git repository at the given path
func Open(path string) (*Repo, error) {
	repo, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		// Init new repo
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}
		repo, err = git.PlainInit(path, false)
		if err != nil {
			return nil, fmt.Errorf("git init: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	return &Repo{path: path, repo: repo}, nil
}

// ReadFile reads a file from the working tree
func (r *Repo) ReadFile(name string) ([]byte, error) {
	path := filepath.Join(r.path, name)
	return os.ReadFile(path)
}

// WriteFile writes a file to the working tree
func (r *Repo) WriteFile(name string, content []byte) error {
	path := filepath.Join(r.path, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

// FileExists reports whether a file exists in the working tree.
func (r *Repo) FileExists(name string) bool {
	path := filepath.Join(r.path, name)
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// Commit creates a git commit with the given message
func (r *Repo) Commit(message, authorName, authorEmail string) error {
	wt, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	// Add all changes
	if _, err := wt.Add("."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are changes to commit
	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if status.IsClean() {
		return fmt.Errorf("no changes to publish: working tree is clean")
	}

	// Commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

// History returns the commit history for a file
func (r *Repo) History(file string, n int) ([]Commit, error) {
	logOptions := &git.LogOptions{}
	if file != "" {
		logOptions.FileName = &file
	}

	iter, err := r.repo.Log(logOptions)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	defer iter.Close()

	var commits []Commit
	count := 0
	iter.ForEach(func(c *object.Commit) error {
		if count >= n {
			return fmt.Errorf("done")
		}
		commits = append(commits, Commit{
			Hash:      c.Hash.String()[:7],
			Message:   strings.TrimSpace(c.Message),
			Author:    c.Author.Name,
			Timestamp: c.Author.When,
		})
		count++
		return nil
	})

	return commits, nil
}

// ListStarFiles returns all .star files in the repo root
func (r *Repo) ListStarFiles() ([]string, error) {
	entries, err := os.ReadDir(r.path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".star") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// ListStarFilesInDir returns all .star files in a repository subdirectory.
func (r *Repo) ListStarFilesInDir(dir string) ([]string, error) {
	path := filepath.Join(r.path, dir)
	entries, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".star") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// EnsureSampleScripts creates sample scripts if they don't exist
func (r *Repo) EnsureSampleScripts() error {
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

	// Check if repo is empty (no commits)
	head, err := r.repo.Head()
	isEmpty := err != nil || head.Hash().IsZero()

	if _, err := r.ReadFile("ios.star"); os.IsNotExist(err) {
		if err := r.WriteFile("ios.star", []byte(iosScript)); err != nil {
			return err
		}
	}
	if _, err := r.ReadFile("android.star"); os.IsNotExist(err) {
		if err := r.WriteFile("android.star", []byte(androidScript)); err != nil {
			return err
		}
	}

	// If repo is empty, make initial commit
	if isEmpty {
		if err := r.Commit("Initial config scripts", "System", "system@kimi.com"); err != nil {
			return err
		}
	}

	return nil
}
