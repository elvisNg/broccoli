package container

import (
	"github.com/elvisNg/broccoli/plugin"
	"github.com/elvisNg/broccoli/plugin/zcontainer"
)

func GetContainer() zcontainer.Container {
	cnt := plugin.NewContainer()
	return cnt
}
