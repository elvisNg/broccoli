package plugin

import (
	"github.com/elvisNg/broccoli/mysql/zmysql"
	"log"
	"net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"github.com/elvisNg/broccoli/config"
	broccolilog "github.com/elvisNg/broccoli/log"
	broccolimongo "github.com/elvisNg/broccoli/mongo"
	"github.com/elvisNg/broccoli/mongo/zmongo"
	broccolimysql "github.com/elvisNg/broccoli/mysql"
	broccoliredis "github.com/elvisNg/broccoli/redis"
	"github.com/elvisNg/broccoli/redis/zredis"
	"github.com/elvisNg/broccoli/sequence"
	tracing "github.com/elvisNg/broccoli/trace"
	"github.com/elvisNg/broccoli/trace/zipkin"
)

// Container contain comm obj, impl zcontainer
type Container struct {
	serviceID     string
	appcfg        config.AppConf
	redis         zredis.Redis
	mongo         zmongo.Mongo
	gomicroClient client.Client
	logger        *logrus.Logger
	tracer        *tracing.TracerWrap
	// http
	httpHandler http.Handler
	// gomicro grpc
	gomicroService micro.Service
	mysql          zmysql.Mysql

	// dbPool          *sql.DB
	// transport       *http.Transport
	// svc             XUtil
	// mqProducer      *mq.MqProducer
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Init(appcfg *config.AppConf) {
	log.Println("[Container.Init] start")
	c.initRedis(&appcfg.Redis)
	c.initLogger(&appcfg.LogConf)
	c.initTracer(&appcfg.Trace)
	c.initMongo(&appcfg.MongoDB)
	c.initMysql(&appcfg.Mysql)
	log.Println("[Container.Init] finish")
	c.appcfg = *appcfg
}

func (c *Container) Reload(appcfg *config.AppConf) {
	log.Println("[Container.Reload] start")
	if c.appcfg.Redis != appcfg.Redis {
		c.reloadRedis(&appcfg.Redis)
	}
	if c.appcfg.LogConf != appcfg.LogConf {
		c.reloadLogger(&appcfg.LogConf)
	}
	if c.appcfg.Trace != appcfg.Trace {
		c.reloadTracer(&appcfg.Trace)
	}
	if c.appcfg.MongoDB != appcfg.MongoDB {
		c.reloadMongo(&appcfg.MongoDB)
	}
	if c.appcfg.Mysql != appcfg.Mysql {
		c.reloadMysql(&appcfg.Mysql)
	}
	log.Println("[Container.Reload] finish")
	c.appcfg = *appcfg
}

// Redis
func (c *Container) initRedis(cfg *config.Redis) {
	if cfg.Enable {
		c.redis = broccoliredis.InitClient(cfg)
	}
}

func (c *Container) reloadRedis(cfg *config.Redis) {
	if cfg.Enable {
		if c.redis != nil {
			c.redis.Reload(cfg)
		} else {
			c.redis = broccoliredis.InitClient(cfg)
		}
	} else if c.redis != nil {
		// 释放
		// c.redis.Release()
		c.redis = nil
	}
}

func (c *Container) GetRedisCli() zredis.Redis {
	return c.redis
}

// Mysql
func (c *Container) initMysql(cfg *config.Mysql) {
	if cfg.Enable {
		c.mysql = broccolimysql.InitClient(cfg)
	}
}

func (c *Container) reloadMysql(cfg *config.Mysql) {
	if cfg.Enable {
		if c.mysql != nil {
			c.mysql.Reload(cfg)
		} else {
			c.mysql = broccolimysql.InitClient(cfg)
		}
	} else if c.mysql != nil {
		// 释放
		// c.mysql.Release()
		c.mysql = nil
	}
}

func (c *Container) GetMyslCli() zmysql.Mysql {
	return c.mysql
}

// GoMicroClient
func (c *Container) SetGoMicroClient(cli client.Client) {
	c.gomicroClient = cli
}

func (c *Container) GetGoMicroClient() client.Client {
	return c.gomicroClient
}

// Logger
func (c *Container) initLogger(cfg *config.LogConf) {
	l, err := broccolilog.New(cfg)
	if err != nil {
		log.Println("initLogger err:", err)
		return
	}
	c.logger = l.Logger
}

func (c *Container) reloadLogger(cfg *config.LogConf) {
	c.initLogger(cfg)
}

func (c *Container) GetLogger() *logrus.Logger {
	return c.logger
}

// func (c *Container) SetDBPool(p *sql.DB) {
// 	c.dbPool = p
// }

// func (c *Container) GetDBPool() *sql.DB {
// 	return c.dbPool
// }

// func (c *Container) SetTransport(tr *http.Transport) {
// 	c.transport = tr
// }

// func (c *Container) GetTransport() *http.Transport {
// 	return c.transport
// }

// func (c *Container) SetSvcOptions(opt interface{}) {
// 	c.serviceOptions = opt
// }

// func (c *Container) GetSvcOptions() interface{} {
// 	return c.serviceOptions
// }

// func (c *Container) SetSvc(svc XUtil) {
// 	c.svc = svc
// }

// func (c *Container) GetSvc() XUtil {
// 	return c.svc
// }

// func (c *Container) SetMQProducer(p *mq.MqProducer) {
// 	c.mqProducer = p
// }

// func (c *Container) GetMQProducer() *mq.MqProducer {
// 	return c.mqProducer
// }

// func (c *Container) Release() {
// if c.redisPool != nil {
// 	c.redisPool.Close()
// }

// if c.dbPool != nil {
// 	c.dbPool.Close()
// }
// }

// Tracer
func (c *Container) initTracer(cfg *config.Trace) (err error) {
	err = zipkin.InitTracer(cfg)
	if err != nil {
		log.Println("initTracer err:", err)
		return
	}
	c.tracer = tracing.NewTracerWrap(opentracing.GlobalTracer())
	return
}

func (c *Container) reloadTracer(cfg *config.Trace) (err error) {
	return c.initTracer(cfg)
}

func (c *Container) GetTracer() *tracing.TracerWrap {
	return c.tracer
}

func (c *Container) SetServiceID(id string) {
	c.serviceID = id
	sequence.Load(id)
}

func (c *Container) GetServiceID() string {
	return c.serviceID
}

func (c *Container) SetHTTPHandler(h http.Handler) {
	c.httpHandler = h
}

func (c *Container) GetHTTPHandler() http.Handler {
	return c.httpHandler
}

func (c *Container) SetGoMicroService(s micro.Service) {
	c.gomicroService = s
}

func (c *Container) GetGoMicroService() micro.Service {
	return c.gomicroService
}

// Mongo
func (c *Container) initMongo(cfg *config.MongoDB) {
	var err error
	if cfg.Enable {
		broccolimongo.InitDefalut(cfg)
		c.mongo, err = broccolimongo.DefaultClient()
		if err != nil {
			log.Println("mgoc.DefaultClient err: ", err)
			return
		}
	}
}

func (c *Container) reloadMongo(cfg *config.MongoDB) {
	var err error
	if cfg.Enable {
		if c.mongo != nil {
			broccolimongo.ReloadDefault(cfg)
			c.mongo, err = broccolimongo.DefaultClient()
			if err != nil {
				log.Println("mgoc.DefaultClient err: ", err)
				return
			}
		} else {
			c.initMongo(cfg)
		}
	} else if c.mongo != nil {
		// 释放
		broccolimongo.DefaultClientRelease()
		c.mongo = nil
	}
}

// GetMongo ...
func (c *Container) GetMongo() zmongo.Mongo {
	return c.mongo
}

//GetMysql
func (c *Container) GetMysql() zmysql.Mysql {
	return c.mysql
}
