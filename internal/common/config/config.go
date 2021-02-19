package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

var rdsCaPath *string

func clowderEnabled() bool {
	return strings.ToLower(os.Getenv("CLOWDER_ENABLED")) != "false"
}

func init() {
	if !clowderEnabled() || clowder.LoadedConfig.Database.RdsCa == nil {
		return
	}

	if rdsCaPathValue, err := clowder.LoadedConfig.RdsCa(); err != nil {
		panic(err)
	} else {
		rdsCaPath = &rdsCaPathValue
	}
}

func Get() *viper.Viper {
	options := viper.New()

	options.SetDefault("openshift.build.commit", "unknown")

	options.SetDefault("log.level", "debug")
	options.SetDefault("demo.mode", false)

	options.SetDefault("http.max.body.size", "512KB")

	options.SetDefault("default.run.timeout", 3600)

	options.SetDefault("db.max.idle.connections", 10)
	options.SetDefault("db.max.open.connections", 20)
	options.SetDefault("migrations.dir", "./migrations")

	options.SetDefault("kafka.timeout", 5000)
	options.SetDefault("kafka.group.id", "playbook-dispatcher")
	options.SetDefault("kafka.auto.offset.reset", "latest")
	options.SetDefault("kafka.auto.commit.interval.ms", 5000)
	options.SetDefault("kafka.request.required.acks", -1) // -1 == "all"
	options.SetDefault("kafka.message.send.max.retries", 15)
	options.SetDefault("kafka.retry.backoff.ms", 100)

	options.SetDefault("schema.message.response", "./schema/playbookRunResponse.message.yaml")
	options.SetDefault("schema.runner.event", "./schema/ansibleRunnerJobEvent.yaml")
	options.SetDefault("schema.api.private", "./schema/private.openapi.yaml")

	options.SetDefault("storage.timeout", 10)
	options.SetDefault("storage.retries", 3)
	options.SetDefault("artifact.max.size", 1024*1024)

	options.SetDefault("cloud.connector.impl", "mock")
	options.SetDefault("cloud.connector.host", "http://cloud-connector:8081")
	options.SetDefault("cloud.connector.timeout", 10)
	options.SetDefault("return.url", "https://cloud.redhat.com/api/v1/ingres/upload")
	options.SetDefault("response.interval", "600")

	options.SetDefault("db.sslmode", "disable")

	if clowderEnabled() {
		options.SetDefault("web.port", clowder.LoadedConfig.WebPort)
		options.SetDefault("metrics.port", clowder.LoadedConfig.MetricsPort)
		options.SetDefault("metrics.path", clowder.LoadedConfig.MetricsPath)

		options.SetDefault("kafka.bootstrap.servers", strings.Join(clowder.KafkaServers, ","))
		options.SetDefault("topic.updates", clowder.KafkaTopics["platform.playbook-dispatcher.runner-updates"].Name)
		options.SetDefault("topic.validation.request", clowder.KafkaTopics["platform.upload.playbook"].Name)
		options.SetDefault("topic.validation.response", clowder.KafkaTopics["platform.upload.validation"].Name)

		options.SetDefault("log.cw.accessKeyId", clowder.LoadedConfig.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("log.cw.secretAccessKey", clowder.LoadedConfig.Logging.Cloudwatch.SecretAccessKey)
		options.SetDefault("log.cw.region", clowder.LoadedConfig.Logging.Cloudwatch.Region)
		options.SetDefault("log.cw.group", clowder.LoadedConfig.Logging.Cloudwatch.LogGroup)

		options.SetDefault("db.host", clowder.LoadedConfig.Database.Hostname)
		options.SetDefault("db.port", clowder.LoadedConfig.Database.Port)
		options.SetDefault("db.name", clowder.LoadedConfig.Database.Name)
		options.SetDefault("db.username", clowder.LoadedConfig.Database.Username)
		options.SetDefault("db.password", clowder.LoadedConfig.Database.Password)

		if rdsCaPath != nil {
			options.SetDefault("db.ca", *rdsCaPath)
		}
	} else {
		options.SetDefault("web.port", 8000)
		options.SetDefault("metrics.port", 9001)
		options.SetDefault("metrics.path", "/metrics")

		options.SetDefault("kafka.bootstrap.servers", "kafka:29092")
		options.SetDefault("topic.updates", "platform.playbook-dispatcher.runner-updates")
		options.SetDefault("topic.validation.request", "platform.upload.playbook")
		options.SetDefault("topic.validation.response", "platform.upload.validation")

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
