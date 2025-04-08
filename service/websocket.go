package service

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket 连接管理
var clients = make(map[uint]*websocket.Conn) // 用户ID到WebSocket连接的映射
var mutex = &sync.Mutex{}                    // 保护 clients 的并发安全

// WebSocket 升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域连接
	},
}

// HandleWebSocket 处理 WebSocket 连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 获取用户ID（假设通过查询参数传递）
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// 转换 userID 为 uint
	var userID uint
	fmt.Sscanf(userIDStr, "%d", &userID)

	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// 将连接加入到 clients
	mutex.Lock()
	clients[userID] = conn
	mutex.Unlock()

	// 监听客户端消息（如果需要）
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			mutex.Lock()
			delete(clients, userID)
			mutex.Unlock()
			break
		}
	}
}

// SendMessageToUser 推送消息给指定用户
func SendMessageToUser(userID uint, message string) {
	mutex.Lock()
	conn, ok := clients[userID]
	mutex.Unlock()

	if ok {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			fmt.Println("WebSocket write error:", err)
			mutex.Lock()
			delete(clients, userID)
			mutex.Unlock()
		}
	}
}
