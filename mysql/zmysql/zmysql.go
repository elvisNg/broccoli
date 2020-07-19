package zmysql

import (
	"github.com/elvisNg/broccoli/config"
	"github.com/jinzhu/gorm"
)

type Mysql interface {
	Reload(cfg *config.Mysql)
	GetCli() *gorm.DB
}
