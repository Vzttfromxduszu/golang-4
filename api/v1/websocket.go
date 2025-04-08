package v1

import (
	"mall/service"

	gin "github.com/gin-gonic/gin"
)

func HandleWebSocket(c *gin.Context) {
	// 处理 WebSocket 连接
	service.HandleWebSocket(c.Writer, c.Request)
}

// 	}
