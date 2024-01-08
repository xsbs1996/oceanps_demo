package logic

import (
	"context"
	"fmt"
	"testing"
	"time"
	"wsdemo/config"
)

// 往 HelloWorld 队列添加数据
func TestRedisPushMsg(t *testing.T) {
	for i := 0; i <= 100; i++ {
		//err := config.OceanpsRedisConf.PushMsgFn(context.Background(), fmt.Sprintf("%s-%s", HelloWorldTopic, ""), []byte("hello world")) // redis
		err := config.OceanpsRabbitMqConf.PushMsgFn(context.Background(), fmt.Sprintf("%s-%s", HelloWorldTopic, ""), []byte("hello world")) // RabbitMQ
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(time.Second * 2)
	}
	return
}
