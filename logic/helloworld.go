package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"wsdemo/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xsbs1996/oceanps"
)

const (
	HelloWorldTopic = "HelloWorld"
)

func init() {
	NewFuncMap[HelloWorldTopic] = NewHelloWorldLogic
	MessageMethodList[fmt.Sprintf("%s.%s", HelloWorldTopic, MessageMethodOperSubscribe)] = MessageMethod{MessageMethodOperSubscribe, HelloWorldTopic}
	MessageMethodList[fmt.Sprintf("%s.%s", HelloWorldTopic, MessageMethodOperUnsubscribe)] = MessageMethod{MessageMethodOperUnsubscribe, HelloWorldTopic}
}

// HelloWorldLogic 通知
type HelloWorldLogic struct {
	Id        int
	Method    string
	ginCtx    *gin.Context
	topic     *oceanps.EventTopic
	topicName string
	conn      *websocket.Conn
	errChan   chan bool
	WriteMsg  chan<- []byte
	ReadMsg   chan []byte
}

func NewHelloWorldLogic(ginCtx *gin.Context, id int, method string, writeMsg chan<- []byte) PubSubLogic {
	logic := &HelloWorldLogic{
		Id:        id,
		Method:    method,
		ginCtx:    ginCtx,
		topic:     nil,
		topicName: HelloWorldTopic,
		errChan:   make(chan bool, 1),
		WriteMsg:  writeMsg,
		ReadMsg:   make(chan []byte, 10),
	}
	return logic
}

func (l *HelloWorldLogic) Process(conn *websocket.Conn, oper string, in interface{}) error {
	switch oper {
	case MessageMethodOperSubscribe:
		return l.subscribe(in, conn)
	case MessageMethodOperUnsubscribe:
		return l.unsubscribe(conn)
	}
	return errors.New("subscribe operation error")
}

// ReadMessage 读取消息
func (l *HelloWorldLogic) ReadMessage(msg []byte) {
	l.ReadMsg <- msg
}

// 订阅消息
func (l *HelloWorldLogic) subscribe(in interface{}, conn *websocket.Conn) error {
	l.setConn(conn)

	// 每个用户是一个主题,全员主题 user 传递空字符串
	l.topic = oceanps.NewEventTopic(l.topicName, "", config.OceanpsRedisConf)

	l.topic.Subscribe(conn, l.WriteMsg)
	// 运行消息发送
	go l.sendMsg()

	return nil
}

// 取消订阅
func (l *HelloWorldLogic) unsubscribe(conn *websocket.Conn) error {
	l.errChan <- true
	l.topic.Unsubscribe(conn, false)
	return nil
}

// 设置链接
func (l *HelloWorldLogic) setConn(conn *websocket.Conn) {
	l.conn = conn
}

// SendMsg 发送消息
func (l *HelloWorldLogic) sendMsg() {
	for {
		select {
		case _ = <-l.ReadMsg: // 接收客户端消息
		case msgBytes, ok := <-l.topic.MsgChan: // 发送消息
			if !ok {
				return
			}

			if l.topic.LenConnList() <= 0 { // 没有接受者直接退出
				return
			}

			// 构建消息体
			msgStruct := &MessageResp{
				Method: l.Method,
				Result: string(msgBytes),
				Id:     l.Id,
				Error:  "",
			}
			msgJson, _ := json.Marshal(msgStruct)

			// 遍历发送消息
			l.topic.ForeachConnList(msgJson)
		case <-l.errChan: // 取消信号
			return
		}
	}
}
