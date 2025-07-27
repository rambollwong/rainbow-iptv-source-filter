package logx

import "github.com/rambollwong/rainbowlog/log"

func InitGlobalLogger() {
	log.DefaultConfigFilePath = "./conf"
	log.DefaultConfigFileName = "config.yaml"
	log.UseDefaultConfigFile()
}
