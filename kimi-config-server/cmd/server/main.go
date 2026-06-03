package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/moonshot-ai/kimi-config-server/internal/api"
	"github.com/moonshot-ai/kimi-config-server/internal/git"
)

func main() {
	repoPath := getEnv("CONFIG_REPO_PATH", "./config-repo")

	// Open or init Git repo
	repo, err := git.Open(repoPath)
	if err != nil {
		log.Fatalf("Failed to open git repo: %v", err)
	}

	// Ensure sample scripts exist
	if err := repo.EnsureSampleScripts(); err != nil {
		log.Fatalf("Failed to ensure sample scripts: %v", err)
	}

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Register routes
	admin := r.Group("/api")
	{
		admin.GET("/platforms", api.ListPlatforms(repo))
		admin.GET("/scripts/:platform", api.GetScript(repo))
		admin.GET("/scripts/:platform/history", api.GetHistory(repo))
		admin.POST("/scripts/:platform", api.SaveScript(repo))
		admin.POST("/scripts/:platform/publish", api.PublishScript(repo))
		admin.POST("/preview", api.PreviewConfig())
	}

	// Client API
	r.GET("/v1/config", api.GetClientConfig(repo))

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
