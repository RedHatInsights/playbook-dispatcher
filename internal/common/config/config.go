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

	options.SetDefault("default.run.timeout", 3600)

	options.SetDefault("db.max.idle.connections", 10)
	options.SetDefault("db.max.open.connections", 20)
	options.SetDefault("migrations.dir", "./migrations")

	options.SetDefault("kafka.group.id", "playbook-dispatcher")
	options.SetDefault("kafka.auto.offset.reset", "latest")
	options.SetDefault("kafka.auto.commit.interval.ms", 5000)
	options.SetDefault("kafka.request.required.acks", -1) // -1 == "all"
	options.SetDefault("kafka.message.send.max.retries", 15)
	options.SetDefault("kafka.retry.backoff.ms", 100)

	options.SetDefault("schema.message.response", "./schema/playbookRunResponse.message.yaml")

	if os.Getenv("CLOWDER_ENABLED") != "false" {
		options.SetDefault("web.port", clowder.LoadedConfig.WebPort)
		options.SetDefault("metrics.port", clowder.LoadedConfig.MetricsPort)
		options.SetDefault("metrics.path", clowder.LoadedConfig.MetricsPath)

		options.SetDefault("kafka.bootstrap.servers", strings.Join(clowder.KafkaServers, ","))
		options.SetDefault("topic.responses", clowder.KafkaTopics["platform.playbook-dispatcher.responses"].Name)

		options.SetDefault("log.cw.accessKeyId", clowder.LoadedConfig.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("log.cw.secretAccessKey", clowder.LoadedConfig.Logging.Cloudwatch.SecretAccessKey)
		options.SetDefault("log.cw.region", clowder.LoadedConfig.Logging.Cloudwatch.Region)
		options.SetDefault("log.cw.group", clowder.LoadedConfig.Logging.Cloudwatch.LogGroup)

		options.SetDefault("db.host", clowder.LoadedConfig.Database.Hostname)
		options.SetDefault("db.port", clowder.LoadedConfig.Database.Port)
		options.SetDefault("db.name", clowder.LoadedConfig.Database.Name)
		options.SetDefault("db.username", clowder.LoadedConfig.Database.Username)
		options.SetDefault("db.password", clowder.LoadedConfig.Database.Password)
	} else {
		options.SetDefault("web.port", 8000)
		options.SetDefault("metrics.port", 9001)
		options.SetDefault("metrics.path", "/metrics")

		options.SetDefault("kafka.bootstrap.servers", "kafka:29092")
		options.SetDefault("topic.responses", "platform.playbook-dispatcher.responses")

		options.SetDefault("db.host", "localhost")
		options.SetDefault("db.port", 5432)
		options.SetDefault("db.name", "insights")
		options.SetDefault("db.username", "insights")
		options.SetDefault("db.password", "insights")
	}

	options.AutomaticEnv()
	options.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return options
}
