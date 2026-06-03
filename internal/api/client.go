package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moonshot-ai/kimi-config-server/internal/git"
	"github.com/moonshot-ai/kimi-config-server/internal/starlark"
)

// GetClientConfig returns the config for a client request
func GetClientConfig(repo *git.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := c.Query("platform")
		version := c.Query("version")
		language := c.Query("lang")
		region := c.Query("region")

		if platform == "" || version == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "platform and version are required"})
			return
		}
		if language == "" {
			language = "en"
		}
		if region == "" {
			region = "domestic"
		}

		// Read script
		scriptBytes, err := repo.ReadFile(platform + ".star")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("config for platform %s not found", platform)})
			return
		}

		// Execute
		ctx := starlark.EvalContext{
			Platform: platform,
			Version:  version,
			Language: language,
			Region:   region,
		}

		result, err := starlark.Execute(string(scriptBytes), ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"config":    result,
			"syncToken": fmt.Sprintf("%s-%s-%s-%s", platform, version, language, region),
		})
	}
}
