package logic

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// NewFuncMap 获取订阅接口函数
var NewFuncMap = map[string]func(ginCtx *gin.Context, id int, method string, writeMsg chan<- []byte) PubSubLogic{}

// PubSubLogic 主题处理接口
type PubSubLogic interface {
	Process(conn *websocket.Conn, oper string, in interface{}) error
	ReadMessage(msg []byte)
	subscribe(in interface{}, conn *websocket.Conn) error
	unsubscribe(conn *websocket.Conn) error
	sendMsg()
	setConn(conn *websocket.Conn)
}

const (
	MessageMethodOperSubscribe   = "Subscribe"   // 订阅
	MessageMethodOperUnsubscribe = "Unsubscribe" // 取消订阅
)

type MessageMethod struct {
	Oper      string // 操作
	TopicName string // 主题
}

var MessageMethodList = map[string]MessageMethod{}

// MessageReq 消息请求参数
type MessageReq struct {
	Method string      `json:"method"` // 请求主题
	Params interface{} `json:"params"` // 请求参数
	Id     int         `json:"id"`     // 请求ID
}

// MessageResp 消息返回参数
type MessageResp struct {
	Method string      `json:"method"` // 请求主题
	Result interface{} `json:"result"` // 回复内容
	Id     int         `json:"id"`     // 请求ID
	Error  string      `json:"error"`  // 错误信息
}
