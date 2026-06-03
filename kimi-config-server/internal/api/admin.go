package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moonshot-ai/kimi-config-server/internal/git"
)

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

		c.JSON(http.StatusOK, gin.H{"platforms": platforms})
	}
}

// GetScript returns a platform script content
func GetScript(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		content, err := repo.ReadFile(platform + ".star")
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

		if err := repo.WriteFile(platform+".star", []byte(req.Content)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "saved"})
	}
}

// GetHistory returns git commit history for a platform
func GetHistory(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Param("platform")
		commits, err := repo.History(platform+".star", 20)
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
