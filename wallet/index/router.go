package index

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) InitRouter() {
	h.GET("/index/inscriptions", func(ctx *gin.Context) {
		fmt.Println("11111111111111111111111111111")
		ctx.JSON(http.StatusOK, gin.H{
			"test": "test",
		})
	})
}
