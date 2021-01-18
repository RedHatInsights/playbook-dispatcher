package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

func Get() *viper.Viper {
	options := viper.New()

	options.SetDefault("openshift.build.commit", "unknown")

	options.SetDefault("log.level", "debug")

	options.SetDefault("http.max.body.size", "512KB")

	if os.Getenv("CLOWDER_ENABLED") != "false" {
		options.SetDefault("web.port", clowder.LoadedConfig.WebPort)
		options.SetDefault("metrics.port", clowder.LoadedConfig.MetricsPort)
		options.SetDefault("metrics.path", clowder.LoadedConfig.MetricsPath)

		options.SetDefault("log.cw.accessKeyId", clowder.LoadedConfig.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("log.cw.secretAccessKey", clowder.LoadedConfig.Logging.Cloudwatch.SecretAccessKey)
		options.SetDefault("log.cw.region", clowder.LoadedConfig.Logging.Cloudwatch.Region)
		options.SetDefault("log.cw.group", clowder.LoadedConfig.Logging.Cloudwatch.LogGroup)
	} else {
		options.SetDefault("web.port", 8080)
		options.SetDefault("metrics.port", 9001)
		options.SetDefault("metrics.path", "/metrics")
	}

	options.AutomaticEnv()
	options.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return options
}
