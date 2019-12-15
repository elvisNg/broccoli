package file

import (
	"github.com/elvisNg/broccoli/config"
	"github.com/elvisNg/broccoli/engine"
)

type ng struct {
}

func (n *ng) Init() (err error) {
	return nil
}

func (n *ng) Subscribe(changes chan interface{}, cancelC chan struct{}) error {
	return nil
}

func (n *ng) GetConfiger() (config.Configer, error) {
	return nil, nil
}

func (n *ng) GetContainer() *engine.Container {
	return nil
}
