package api

import (
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moonshot-ai/kimi-config-server/internal/git"
)

type platformVersion struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Latest  bool   `json:"latest"`
	Legacy  bool   `json:"legacy"`
	Draft   bool   `json:"draft"`
}

var versionFilePattern = regexp.MustCompile(`^v([0-9]+)\.star$`)

// ListPlatforms returns available platform scripts
func ListPlatforms(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		files, err := repo.ListStarFiles()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		platforms := make([]string, 0, len(files))
		for _, f := range files {
			platforms = append(platforms, strings.TrimSuffix(f, ".star"))
		}

		for _, platform := range []string{"ios", "android"} {
			versionFiles, err := repo.ListStarFilesInDir(platform)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if len(versionFiles) > 0 && !contains(platforms, platform) {
				platforms = append(platforms, platform)
			}
		}

		sort.Strings(platforms)
		c.JSON(http.StatusOK, gin.H{"platforms": platforms})
	}
}

// GetScript returns a platform script content
func GetScript(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		file := scriptFileForRequest(repo, platform, c.Query("version"))
		content, err := repo.ReadFile(file)
		if err != nil {
			if strings.Contains(err.Error(), "no such file") {
				c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"platform": platform,
			"path":     file,
			"version":  c.Query("version"),
			"content":  string(content),
		})
	}
}

// SaveScript saves a script to the working tree (without committing)
type saveScriptReq struct {
	Content string `json:"content" binding:"required"`
}

func SaveScript(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		var req saveScriptReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		file := scriptFileForWrite(platform, c.Query("version"))
		if err := repo.WriteFile(file, []byte(req.Content)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "saved", "path": file})
	}
}

// ListVersions returns versioned scripts for a platform and a next draft version.
func ListVersions(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		versions, err := listPlatformVersions(repo, platform)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		latest := latestVersion(versions)
		next := "v1"
		if latest != nil {
			next = incrementVersion(latest.Version)
		}

		c.JSON(http.StatusOK, gin.H{
			"platform":    platform,
			"versions":    versions,
			"latest":      latest,
			"nextVersion": next,
		})
	}
}

// CreateDraft copies the latest platform script into an in-memory draft response.
func CreateDraft(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		versions, err := listPlatformVersions(repo, platform)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		latest := latestVersion(versions)
		if latest == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "latest script not found"})
			return
		}

		next := incrementVersion(latest.Version)
		target := versionedScriptFile(platform, next)
		content, err := repo.ReadFile(latest.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"platform":       platform,
			"version":        next,
			"path":           target,
			"content":        string(content),
			"baseVersion":    latest.Version,
			"basePath":       latest.Path,
			"baseContent":    string(content),
			"latestVersion":  latest.Version,
			"workingVersion": next,
		})
	}
}

// GetHistory returns git commit history for a platform
func GetHistory(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		commits, err := repo.History(scriptFileForRequest(repo, platform, c.Query("version")), 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"platform": platform,
			"commits":  commits,
		})
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func listPlatformVersions(repo *git.Repo, platform string) ([]platformVersion, error) {
	files, err := repo.ListStarFilesInDir(platform)
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

	if repo.FileExists(platform+".star") && !hasVersion(versions, "v1") {
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

func scriptFileForRequest(repo *git.Repo, platform, version string) string {
	if version == "" {
		versions, err := listPlatformVersions(repo, platform)
		if err == nil {
			if latest := latestVersion(versions); latest != nil {
				return latest.Path
			}
		}
		return platform + ".star"
	}

	file := versionedScriptFile(platform, version)
	if repo.FileExists(file) {
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

// PublishScript commits the script to git
type publishReq struct {
	Message string `json:"message"`
	Author  string `json:"author"`
	Email   string `json:"email"`
}

func PublishScript(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		var req publishReq
		c.ShouldBindJSON(&req)

		message := req.Message
		if message == "" {
			message = "Update " + platform + " config"
		}
		author := req.Author
		if author == "" {
			author = "Anonymous"
		}
		email := req.Email
		if email == "" {
			email = "anonymous@kimi.com"
		}

		if err := repo.Commit(message, author, email); err != nil {
			if strings.Contains(err.Error(), "no changes to publish") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No changes to publish. Please edit and save the script first."})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "published"})
	}
}
