package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wsdemo/logic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ServeWs websocket链接
func ServeWs() gin.HandlerFunc {
	return func(c *gin.Context) {
		client, err := pool.createConnection(c.Writer, c.Request, c)
		if err != nil {
			panic(err)
			return
		}

		// 读处理
		go client.read()

		// 写处理
		go client.write()
	}
}

// 读取操作
func (client *WsClient) read() {
	defer client.close()

	client.conn.SetReadLimit(maxMessageSize)
	// 设置连接的读超时时间
	err := client.conn.SetReadDeadline(time.Now().Add(readWait))
	if err != nil {
		panic(err)
		return
	}

	// 读取客户端发送的消息
readPumpEvent:
	for {
		// 读取客户端发送的消息
		messageType, message, err := client.conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.TextMessage {
			continue readPumpEvent
		}

		// 充值读写等待时间
		if err := client.conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
			continue readPumpEvent
		}
		if err := client.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			continue readPumpEvent
		}

		// ping-pong
		if string(message) == PingMsg {
			client.WriteMsg <- []byte(PongMsg)
			continue readPumpEvent
		}

		fmt.Println(string(message))
		// 解析message
		var req = &logic.MessageReq{}
		if err := json.Unmarshal(message, &req); err != nil {
			continue readPumpEvent
		}

		fmt.Println(logic.NewFuncMap)
		fmt.Println(logic.MessageMethodList)
		// 查看请求主题是否存在
		method, ok := logic.MessageMethodList[req.Method]
		if !ok {
			client.writeErr(req.Method, req.Id, errors.New("topic does not exist"))
			continue readPumpEvent
		}

		// 查看是否已订阅
		switch method.Oper {
		case logic.MessageMethodOperSubscribe: // 订阅操作
			// 查看是否已订阅
			wsLogic, ok := client.isTopic(method.TopicName)
			if ok {
				wsLogic.ReadMessage(message)
				continue readPumpEvent
			}

			// 未订阅则订阅
			fn, ok := logic.NewFuncMap[method.TopicName]
			if !ok {
				client.writeErr(req.Method, req.Id, errors.New("topic func does not exist"))
				continue readPumpEvent
			}
			l := fn(client.ginCtx, req.Id, req.Method, client.WriteMsg)
			operErr := l.Process(client.conn, method.Oper, req.Params)
			if operErr != nil {
				client.errTotal += 1
				if client.errTotal >= wsClientMaxErrTotal { // 订阅失败达到一定次数，直接链接关闭
					return
				}

				// 向客户端发送订阅失败的消息
				client.writeErr(req.Method, req.Id, errors.New("subject subscription failed"))
				continue
			}

			// 添加主题，发送消息
			client.addTopic(method.TopicName, l)
			l.ReadMessage(message)
		case logic.MessageMethodOperUnsubscribe: // 取消订阅
			l, ok := client.isTopic(method.TopicName)
			if !ok {
				continue
			}
			_ = l.Process(client.conn, method.Oper, req.Params)
			client.delTopic(method.TopicName)
		}
	}
}

// 写入操作
func (client *WsClient) write() {
	defer client.close()

	for {
		select {
		case message, ok := <-client.WriteMsg:
			if !ok {
				return
			}

			// 充值读写等待时间
			if err := client.conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
				continue
			}
			if err := client.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				continue
			}

			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// 写入消息
			if _, err := w.Write(message); err != nil {
				return
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-client.exitErr:
			return
		}
	}
}

// writeErr 写入失败消息
func (client *WsClient) writeErr(method string, id int, err error) {
	// 向客户端发送订阅失败的消息
	msg := &logic.MessageResp{
		Method: method,
		Result: nil,
		Id:     id,
		Error:  err.Error(),
	}
	msgJson, _ := json.Marshal(msg)
	client.WriteMsg <- msgJson
}

// 查看是否已订阅
func (client *WsClient) isTopic(topicName string) (logic.PubSubLogic, bool) {
	client.mx.Lock()
	defer client.mx.Unlock()
	l, ok := client.logicList[topicName]
	return l, ok
}

// 添加订阅
func (client *WsClient) addTopic(topicName string, l logic.PubSubLogic) {
	client.mx.Lock()
	defer client.mx.Unlock()
	client.logicList[topicName] = l
	return
}

// 删除订阅
func (client *WsClient) delTopic(topicName string) {
	client.mx.Lock()
	defer client.mx.Unlock()
	delete(client.logicList, topicName)
	return
}

// 订阅数量
func (client *WsClient) lenTopic() int {
	client.mx.Lock()
	defer client.mx.Unlock()
	return len(client.logicList)
}

// 链接退出
func (client *WsClient) close() {
	client.mx.Lock()
	defer client.mx.Unlock()
	if client.exitBool {
		return
	}

	pool.delConnection(client.conn)
	_ = client.conn.Close()
	client.exitErr <- struct{}{}
	for _, l := range client.logicList {
		_ = l.Process(client.conn, logic.MessageMethodOperUnsubscribe, nil)
	}
	client.exitBool = true
}
