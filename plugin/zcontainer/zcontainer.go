package zcontainer

import (
	"github.com/elvisNg/broccoli/mysql/zmysql"
	"net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"

	"github.com/elvisNg/broccoli/config"
	"github.com/elvisNg/broccoli/mongo/zmongo"
	"github.com/elvisNg/broccoli/redis/zredis"
	tracing "github.com/elvisNg/broccoli/trace"
)

// Container 组件的容器访问接口
type Container interface {
	Init(appcfg *config.AppConf)
	Reload(appcfg *config.AppConf)
	GetRedisCli() zredis.Redis
	SetGoMicroClient(cli client.Client)
	GetGoMicroClient() client.Client
	GetLogger() *logrus.Logger
	GetTracer() *tracing.TracerWrap
	SetServiceID(id string)
	GetServiceID() string
	SetHTTPHandler(h http.Handler)
	GetHTTPHandler() http.Handler
	SetGoMicroService(s micro.Service)
	GetGoMicroService() micro.Service
	GetMongo() zmongo.Mongo
	GetMysql() zmysql.Mysql
}
