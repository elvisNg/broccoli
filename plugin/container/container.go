package container

import (
	"github.com/elvisNg/broccoli/plugin"
)

func GetContainer() *plugin.Container {
	cnt := plugin.NewContainer()
	return cnt
}
