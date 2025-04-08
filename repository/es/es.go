package es

import (
	"log"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v7"

	"mall/conf"
)

var EsClient *elastic.Client

// InitEs 初始化es
func InitEs() {
	esConn := "http://" + conf.EsHost + ":" + conf.EsPort
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(esConn))
	if err != nil {
		log.Panic("Failed to initialize Elasticsearch client:", err)
	}
	EsClient = client
}

// EsHookLog 初始化log日志
func EsHookLog() *elogrus.ElasticHook {
	if EsClient == nil {
		log.Println("Elasticsearch client is not initialized")
		return nil
	}
	// 将 Logrus 的日志发送到 Elasticsearch
	hook, err := elogrus.NewElasticHook(EsClient, conf.EsHost, logrus.DebugLevel, conf.EsIndex)
	if err != nil {
		log.Panic("Failed to create Elasticsearch hook:", err)
	}
	return hook
}
