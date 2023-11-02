package main

import (
	"wsdemo/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建Gin应用
	app := gin.Default()

	// 注册WebSocket路由
	app.GET("/wss", handler.ServeWs())

	// 启动应用
	err := app.Run(":8080")
	if err != nil {
		panic(err)
	}
}
