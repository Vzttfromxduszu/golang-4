package main

import (
	"mall/conf"
	"mall/loading"
	util "mall/pkg/utils"
	"mall/pkg/utils/track"

	// "mall/repository/es"
	"mall/repository/mq"
	"mall/routes"
	"mall/service"
	"os"
	"runtime/trace"
)

func main() {
	// Ek1+Ep1==Ek2+Ep2
	conf.Init()
	loading.Loading()
	f, _ := os.Create("trace1.out")
	defer f.Close()
	trace.Start(f)
	defer trace.Stop()
	// es.InitEs() // 初始化es
	mq.InitRabbitMQ()
	track.InitJaeger()
	defer func() {
		if mq.RabbitMQ != nil {
			mq.RabbitMQ.Close()
		}
	}()
	// 启动消费者
	go func() {
		util.LogrusObj.Infoln("Start to consume skill goods")
		service.ConsumeSecKillGoods()
	}()
	r := routes.NewRouter()
	_ = r.Run(conf.HttpPort)

}
