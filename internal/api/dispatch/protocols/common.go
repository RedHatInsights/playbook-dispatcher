package protocols

import "github.com/spf13/viper"

func buildCommonSignal(cfg *viper.Viper) map[string]string {
	return map[string]string{
		"return_url":        cfg.GetString("return.url"),
		"response_interval": cfg.GetString("response.interval"),
	}
}
