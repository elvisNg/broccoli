package generator

func GenerateGlobal(PD *Generator, rootdir string) (err error) {
	err = genGlobal(PD, rootdir)
	if err != nil {
		return
	}
	err = genGlobalInit(PD, rootdir)
	if err != nil {
		return
	}

	return
}

func genGlobalInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	context := `package global

import (
	"log"

	"github.com/elvisNg/broccoli/config"
	"github.com/elvisNg/broccoli/engine"
	"github.com/elvisNg/broccoli/service"
)

var ng engine.Engine
var ServiceOpts []service.Option

func init() {
	// load engine
	loadEngineFnOpt := service.WithLoadEngineFnOption(func(ng engine.Engine) {
		log.Println("WithLoadEngineFnOption: SetNG success.")
		SetNG(ng)
	})
	ServiceOpts = append(ServiceOpts, loadEngineFnOpt)
	// // server wrap
	// ServiceOpts = append(ServiceOpts, service.WithGoMicroServerWrapGenerateFnOption(gomicro.GenerateServerLogWrap))
}

// GetNG ...
func GetNG() engine.Engine {
	return ng
}

// SetNG ...
func SetNG(n engine.Engine) {
	ng = n
}

// GetConfig ...
func GetConfig() (conf *config.AppConf) {
	c, err := ng.GetConfiger()
	if err != nil {
		log.Println("global.GetConfig err:", err)
		return
	}
	conf = c.Get()
	return
}

`
	fn := GetTargetFileName(PD, "global.init", rootdir)
	return writeContext(fn, header, context, false)
}

func genGlobal(PD *Generator, rootdir string) error {
	header := ``
	context := `package global

`
	fn := GetTargetFileName(PD, "global", rootdir)
	return writeContext(fn, header, context, false)
}
