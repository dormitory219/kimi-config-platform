package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moonshot-ai/kimi-config-server/internal/starlark"
)

// PreviewConfig executes a script with given context and returns the result
type previewReq struct {
	Script string              `json:"script" binding:"required"`
	Ctx    starlark.EvalContext `json:"ctx" binding:"required"`
}

func PreviewConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req previewReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := starlark.Execute(req.Script, req.Ctx)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"config": result,
		})
	}
}
