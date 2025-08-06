package conf

import (
	"path"

	"github.com/rambollwong/rainbow-iptv-source-filter/pkg/proto"
	"github.com/rambollwong/rainbowlog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	Config *proto.Config // Global config

	_ = pflag.StringP("config.path", "c", "./conf", "config file path")
	_ = pflag.StringP("local-path", "l", "", "path of local program list source file")
	_ = pflag.StringP("output", "o", "", "output file path")
)

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

	Config = &proto.Config{}
	if err := viper.Unmarshal(Config); err != nil {
		return err
	}
	if localPath != "" {
		Config.ProgramListSourceFileLocalPath = path.Join(localPath)
	}
	if outputPath != "" {
		Config.OutputFile = path.Join(outputPath)
	}

	return nil
}
