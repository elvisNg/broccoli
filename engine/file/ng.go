package file

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"time"

	"github.com/elvisNg/broccoli/config"
	"github.com/elvisNg/broccoli/engine"
	"github.com/elvisNg/broccoli/plugin/zcontainer"
	"github.com/elvisNg/broccoli/utils"
)

// ng fileengine
type ng struct {
	entry                *config.Entry
	configer             config.Configer
	prevRawConfigContent []byte
	container            zcontainer.Container
	context              context.Context
	cancelFunc           context.CancelFunc
	options              *Options
}

type Options struct {
	context context.Context
}

type Option func(o *Options)

func New(entry *config.Entry, container zcontainer.Container, opts ...Option) (engine.Engine, error) {
	n := &ng{
		entry:     entry,
		container: container,
		options:   &Options{},
	}
	for _, o := range opts {
		o(n.options)
	}

	return n, nil
}

func (n *ng) Init() (err error) {
	// ��ȡ����
	d, err := ioutil.ReadFile(n.entry.ConfigPath)
	if err != nil {
		return
	}

	if err = n.refreshConfig(d); err != nil {
		return
	}

	return nil
}

func (n *ng) Subscribe(changes chan interface{}, cancelC chan struct{}) error {
	for {
		time.Sleep(10 * time.Second)

		d, err := ioutil.ReadFile(n.entry.ConfigPath)
		if err != nil {
			log.Printf("[zeus] [engine.Subscribe] error: %s\n", err)
			continue
		}
		// TODO: ����ļ������޸�
		if string(d) == "" || string(n.prevRawConfigContent) == string(d) {
			continue
		}
		if err = n.refreshConfig(d); err != nil {
			log.Printf("[zeus] [engine.Subscribe] ignore '%s', error: %s\n", string(d), err)
			continue
		}
		if n.configer != nil {
			log.Printf("[zeus] [engine.Subscribe] configPath: %s change\n", n.entry.ConfigPath)
			select {
			case changes <- n.configer:
			case <-cancelC:
				log.Printf("[zeus] [engine.Subscribe] cancel watch config: %s\n", n.entry.ConfigPath)
				return nil
			default: // ��ֹ��������changes����һֱ����
				log.Printf("[zeus] [engine.Subscribe] channel is blocked, can not push change into changes")
			}
		}
	}
	return nil
}

func (n *ng) GetConfiger() (config.Configer, error) {
	return n.configer, nil
}

func (n *ng) GetContainer() zcontainer.Container {
	return n.container
}

// refreshConfig ˢ�����ã�ʧ������ԭ�����ã���Ӱ�쵱ǰ������
func (n *ng) refreshConfig(content []byte) (err error) {
	log.Printf("[zeus] [engine.refreshConfig] configpath: %s��configcontent: %s\n", n.entry.ConfigPath, string(content))
	configFormat := n.entry.ConfigFormat
	if utils.IsEmptyString(configFormat) {
		configFormat = "json"
	}
	var configer config.Configer
	switch configFormat {
	case "json":
		jsoner := &config.Jsoner{}
		err = jsoner.Init(content)
		if err != nil {
			msg := "[zeus] [engine.refreshConfig] jsoner ��������ʧ�ܣ�configpath: " + n.entry.ConfigPath
			log.Println(msg)
			return
		}
		configer = jsoner
	case "toml":
		msg := "[zeus] [engine.refreshConfig] toml:��֧�ֵ����ø�ʽ��configpath: " + n.entry.ConfigPath
		log.Println(msg)
		err = errors.New(msg)
		return
	default:
		msg := "[zeus] [engine.refreshConfig] " + configFormat + ":��֧�ֵ����ø�ʽ��configpath: " + n.entry.ConfigPath
		log.Println(msg)
		err = errors.New(msg)
		return
	}

	if configer != nil {
		log.Printf("[zeus] [engine.refreshConfig] ˢ�����óɹ���configpath: %s\n", n.entry.ConfigPath)
		n.configer = configer
		n.prevRawConfigContent = content
		return
	}
	log.Printf("[zeus] [engine.refreshConfig] ˢ������ʧ�ܣ�����ԭ�����ã�configpath: %s\n", n.entry.ConfigPath)
	return
}
