package conf

import (
	"path"
	"strings"

	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
	"github.com/rambollwong/rainbowlog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	Config *config // Global config

	_ = pflag.StringP("config.path", "c", "./conf", "config file path")
	_ = pflag.StringP("local-path", "l", "", "path of local program list source file")
	_ = pflag.StringP("output", "o", "", "output file path")
)

type config struct {
	*proto.Config

	HostCustomUA map[string]string
}

func InitConfig() error {
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return err
	}
	configPath := viper.GetString("config.path")
	log.Info().Msg("load config file").Str("config.path", configPath).Done()
	localPath := viper.GetString("local-path")
	outputPath := viper.GetString("output")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	if err := viper.MergeInConfig(); err != nil {
		return err
	}

	pConf := &proto.Config{}
	if err := viper.Unmarshal(pConf); err != nil {
		return err
	}
	if localPath != "" {
		pConf.ProgramListSourceFileLocalPath = path.Join(localPath)
	}
	if outputPath != "" {
		pConf.OutputFile = path.Join(outputPath)
	}

	Config = &config{
		Config:       pConf,
		HostCustomUA: make(map[string]string, len(pConf.HostCustomUA)),
	}

	for _, s := range pConf.HostCustomUA {
		arr := strings.Split(strings.TrimSpace(s), "->")
		if len(arr) != 2 {
			log.Warn().Msg("invalid host_custom_ua, ignore.").Str("host_custom_ua", s).Done()
			continue
		}
		Config.HostCustomUA[strings.TrimSpace(arr[0])] = strings.TrimSpace(arr[1])
	}

	return nil
}
