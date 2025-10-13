package logx

import (
	"github.com/rambollwong/rainbowlog"
	"github.com/rambollwong/rainbowlog/log"
)

func InitGlobalLogger(configFilePath string) {
	if configFilePath != "" {
		log.DefaultConfigFilePath = configFilePath
	} else {
		log.DefaultConfigFilePath = "./conf"
	}
	log.DefaultConfigFileName = "config.yaml"
	log.UseDefaultConfigFile(rainbowlog.WithCallerMarshalFunc(rainbowlog.NewShortCallerMarshalFunc("rainbow-iptv-source-filter")))
}
