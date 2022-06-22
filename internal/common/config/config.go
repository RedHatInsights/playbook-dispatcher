package config

import (
	"strings"

	"github.com/spf13/viper"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

var rdsCaPath *string

func init() {
	if clowder.IsClowderEnabled() || clowder.LoadedConfig.Database.RdsCa == nil {
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

	options.SetDefault("build.commit", "unknown")

	options.SetDefault("log.level", "debug")
	options.SetDefault("demo.mode", false)

	options.SetDefault("http.max.body.size", "512KB")

	options.SetDefault("default.run.timeout", 3600)

	options.SetDefault("db.max.idle.connections", 10)
	options.SetDefault("db.max.open.connections", 20)
	options.SetDefault("migrations.dir", "./migrations")

	options.SetDefault("kafka.timeout", 10000)
	options.SetDefault("kafka.group.id", "playbook-dispatcher")
	options.SetDefault("kafka.auto.offset.reset", "latest")
	options.SetDefault("kafka.auto.commit.interval.ms", 5000)
	options.SetDefault("kafka.request.required.acks", -1) // -1 == "all"
	options.SetDefault("kafka.message.send.max.retries", 15)
	options.SetDefault("kafka.retry.backoff.ms", 100)

	options.SetDefault("schema.message.response", "./schema/playbookRunResponse.message.yaml")
	options.SetDefault("schema.satmessage.response", "./schema/playbookSatRunResponse.message.yaml")
	options.SetDefault("schema.runner.event", "./schema/ansibleRunnerJobEvent.yaml")
	options.SetDefault("schema.rhcsat.event", "./schema/rhcsatJobEvent.yaml")
	options.SetDefault("schema.api.private", "./schema/private.openapi.yaml")

	options.SetDefault("storage.timeout", 10)
	options.SetDefault("storage.retries", 3)
	options.SetDefault("storage.max.concurrency", 5)
	options.SetDefault("artifact.max.size", 1024*1024)

	options.SetDefault("satellite.response.full", true)

	options.SetDefault("cloud.connector.impl", "mock")
	options.SetDefault("cloud.connector.host", "cloud-connector")
	options.SetDefault("cloud.connector.port", "8080")
	options.SetDefault("cloud.connector.scheme", "http")
	options.SetDefault("cloud.connector.timeout", 10)
	options.SetDefault("cloud.connector.client.id", "playbook-dispatcher")
	options.SetDefault("cloud.connector.psk", "")
	options.SetDefault("cloud.connector.rps", 100)
	options.SetDefault("cloud.connector.req.bucket", 60)

	options.SetDefault("return.url", "https://cloud.redhat.com/api/ingress/v1/upload")
	options.SetDefault("web.console.url.default", "https://console.redhat.com")
	options.SetDefault("response.interval", "600")

	options.SetDefault("rbac.impl", "mock")
	options.SetDefault("rbac.host", "rbac")
	options.SetDefault("rbac.port", "8080")
	options.SetDefault("rbac.scheme", "http")
	options.SetDefault("rbac.timeout", 10)

	options.SetDefault("tenant.translator.impl", "dynamic-mock")
	options.SetDefault("tenant.translator.host", "localhost")
	options.SetDefault("tenant.translator.scheme", "http")
	options.SetDefault("tenant.translator.port", "8892")
	options.SetDefault("tenant.translator.timeout", 10)

	options.SetDefault("db.sslmode", "disable")

	if clowder.IsClowderEnabled() {

		cfg := clowder.LoadedConfig
		broker := cfg.Kafka.Brokers[0]

		options.SetDefault("web.port", cfg.PublicPort)
		options.SetDefault("metrics.port", cfg.MetricsPort)
		options.SetDefault("metrics.path", cfg.MetricsPath)

		options.SetDefault("kafka.bootstrap.servers", strings.Join(clowder.KafkaServers, ","))
		options.SetDefault("topic.updates", clowder.KafkaTopics["platform.playbook-dispatcher.runner-updates"].Name)
		options.SetDefault("topic.validation.request", clowder.KafkaTopics["platform.upload.announce"].Name)
		options.SetDefault("topic.validation.response", clowder.KafkaTopics["platform.upload.validation"].Name)

		if broker.Authtype != nil {
			options.Set("kafka.sasl.username", *broker.Sasl.Username)
			options.Set("kafka.sasl.password", *broker.Sasl.Password)
			options.Set("kafka.sasl.mechanism", *broker.Sasl.SaslMechanism)
			options.Set("kafka.sasl.protocol", *broker.Sasl.SecurityProtocol)
			caPath, err := cfg.KafkaCa(broker)
			if err != nil {
				panic("Kafka CA failed to write")
			}
			options.Set("kafka.capath", caPath)
		}

		options.SetDefault("log.cw.accessKeyId", cfg.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("log.cw.secretAccessKey", cfg.Logging.Cloudwatch.SecretAccessKey)
		options.SetDefault("log.cw.region", cfg.Logging.Cloudwatch.Region)
		options.SetDefault("log.cw.group", cfg.Logging.Cloudwatch.LogGroup)

		options.SetDefault("db.host", cfg.Database.Hostname)
		options.SetDefault("db.port", cfg.Database.Port)
		options.SetDefault("db.name", cfg.Database.Name)
		options.SetDefault("db.username", cfg.Database.Username)
		options.SetDefault("db.password", cfg.Database.Password)

		if rdsCaPath != nil {
			options.SetDefault("db.ca", *rdsCaPath)
		}
	} else {
		options.SetDefault("web.port", 8000)
		options.SetDefault("metrics.port", 9001)
		options.SetDefault("metrics.path", "/metrics")

		options.SetDefault("kafka.bootstrap.servers", "kafka:29092")
		options.SetDefault("topic.updates", "platform.playbook-dispatcher.runner-updates")
		options.SetDefault("topic.validation.request", "platform.upload.announce")
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
