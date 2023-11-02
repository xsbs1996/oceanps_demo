package handler

import (
	"sync"
	"time"

	"wsdemo/logic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait           = 20 * time.Second // 往客户端传送消息的等待时间
	readWait            = 20 * time.Second // 从客户端获取消息的等待时间
	maxMessageSize      = 1024             // 允许的最大消息长度
	wsClientMaxErrTotal = 20               // 错误链接允许的重试次数
	PingMsg             = "ping"           // ping
	PongMsg             = "pong"           // pong
)

type WsClient struct {
	mx        sync.Mutex                   // 锁
	ginCtx    *gin.Context                 // gin上下文
	conn      *websocket.Conn              // 链接
	logicList map[string]logic.PubSubLogic // 主题处理程序
	WriteMsg  chan []byte                  // 写队列
	exitErr   chan struct{}                // 退出信号
	exitBool  bool                         // 是否已退出
	errTotal  int                          // 失败次数
}

// ConnectionPool 链接池
type ConnectionPool struct {
	mx       sync.RWMutex                  // 读写锁
	pool     map[*websocket.Conn]*WsClient // 链接池
	total    int                           // 总数
	maxTotal int                           // 最大数
}
