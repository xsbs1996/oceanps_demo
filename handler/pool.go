package handler

import (
	"errors"
	"net/http"
	"sync"

	"wsdemo/logic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var pool *ConnectionPool

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	pool = &ConnectionPool{
		mx:       sync.RWMutex{},
		pool:     make(map[*websocket.Conn]*WsClient),
		total:    0,
		maxTotal: 1000,
	}
}

// createConnection 创建连接
func (cp *ConnectionPool) createConnection(w http.ResponseWriter, r *http.Request, ginCtx *gin.Context) (*WsClient, error) {
	cp.mx.Lock()
	defer cp.mx.Unlock()

	// 大于链接最大数辆
	if cp.total > cp.maxTotal {
		return nil, errors.New("too many connections")
	}

	// 升级链接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	// 放入链接
	cp.pool[conn] = &WsClient{
		mx:        sync.Mutex{},
		ginCtx:    ginCtx,
		conn:      conn,
		WriteMsg:  make(chan []byte, 20),
		exitErr:   make(chan struct{}, 1),
		logicList: make(map[string]logic.PubSubLogic),
	}

	cp.total = len(cp.pool)
	return cp.pool[conn], nil
}

// delConnection 删除连接
func (cp *ConnectionPool) delConnection(conn *websocket.Conn) {
	cp.mx.Lock()
	defer cp.mx.Unlock()
	if ws, ok := cp.pool[conn]; ok {
		for _, v := range ws.logicList { // 取消订阅
			_ = v.Process(conn, logic.MessageMethodOperUnsubscribe, nil)
		}

		delete(cp.pool, conn)
		cp.total = len(cp.pool)
	}
}
