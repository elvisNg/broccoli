package zredis

import (
	"github.com/elvisNg/broccoli/config"
	"github.com/go-redis/redis"
)

type Redis interface {
	Reload(cfg *config.Redis)
	GetCli() *redis.Client
}
