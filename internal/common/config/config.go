package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

const (
	// Kessel authorization modes (matching Unleash feature flag variants)
	KesselModeRBACOnly           = "rbac-only"
	KesselModeBothRBACEnforces   = "both-rbac-enforces"
	KesselModeBothKesselEnforces = "both-kessel-enforces"
	KesselModeKesselOnly         = "kessel-only"
)

var rdsCaPath *string

func init() {
	if !clowder.IsClowderEnabled() || clowder.LoadedConfig.Database.RdsCa == nil {
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
	options.SetDefault("artifact.truncate.stdout.field.after.lines", 500)
	options.SetDefault("artifact.max.stdout.field.size", 1024)
	options.SetDefault("artifact.max.kafka.message.size", 1024*1024)

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
	options.SetDefault("rbac.timeout", 60)

	options.SetDefault("inventory.connector.impl", "mock")
	options.SetDefault("inventory.connector.host", "localhost")
	options.SetDefault("inventory.connector.port", "4000")
	options.SetDefault("inventory.connector.scheme", "http")
	options.SetDefault("inventory.connector.ordered.by", "display_name")
	options.SetDefault("inventory.connector.ordered.how", "ASC")
	options.SetDefault("inventory.connector.limit", 100)
	options.SetDefault("inventory.connector.offset", 0)
	options.SetDefault("inventory.connector.timeout", 10)

	options.SetDefault("sources.impl", "mock")
	options.SetDefault("sources.host", "sources")
	options.SetDefault("sources.port", "8080")
	options.SetDefault("sources.scheme", "http")
	options.SetDefault("sources.timeout", 10)

	options.SetDefault("tenant.translator.impl", "dynamic-mock")
	options.SetDefault("tenant.translator.host", "localhost")
	options.SetDefault("tenant.translator.scheme", "http")
	options.SetDefault("tenant.translator.port", "8892")
	options.SetDefault("tenant.translator.timeout", 10)

	options.SetDefault("db.sslmode", "disable")

	options.SetDefault("blocklist.org.ids", "")

	// Kessel authorization configuration
	// Feature flag: master switch for Kessel authorization
	options.SetDefault("kessel.enabled", false)
	// Feature flag: authorization mode matching Unleash variants
	// Valid values: rbac-only, both-rbac-enforces, both-kessel-enforces, kessel-only
	options.SetDefault("kessel.auth.mode", "rbac-only")

	// Kessel client configuration
	options.SetDefault("kessel.url", "localhost:9091")
	options.SetDefault("kessel.auth.enabled", false)
	options.SetDefault("kessel.auth.client.id", "")
	options.SetDefault("kessel.auth.client.secret", "")
	options.SetDefault("kessel.auth.oidc.issuer", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token")
	options.SetDefault("kessel.insecure", true)

	// Unleash feature flag configuration (defaults for non-Clowder environments)
	options.SetDefault("unleash.enabled", false)
	options.SetDefault("unleash.url", "")
	options.SetDefault("unleash.api.token", "")
	options.SetDefault("unleash.app.name", "playbook-dispatcher")
	options.SetDefault("unleash.environment", "development")

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
		}
		if broker.Cacert != nil {
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

		// Unleash (Feature Flags) configuration from Clowder
		// Clowder provides this in stage/production environments
		if cfg.FeatureFlags != nil {
			unleashURL := ""
			if cfg.FeatureFlags.Hostname != "" && cfg.FeatureFlags.Port != 0 && cfg.FeatureFlags.Scheme != "" {
				unleashURL = fmt.Sprintf("%s://%s:%d/api",
					cfg.FeatureFlags.Scheme,
					cfg.FeatureFlags.Hostname,
					cfg.FeatureFlags.Port)
			}

			if unleashURL != "" {
				options.SetDefault("unleash.url", unleashURL)
				options.SetDefault("unleash.enabled", true)
			}

			if cfg.FeatureFlags.ClientAccessToken != nil {
				options.SetDefault("unleash.api.token", *cfg.FeatureFlags.ClientAccessToken)
			}
		}

		// Kessel endpoint discovery from Clowder
		// Currently commented out due to RHCLOUD-40314
		// Uncomment when Kessel inventory is properly registered in Clowder
		/*
			for _, e := range cfg.Endpoints {
				if e.App == "kessel-inventory-api" {
					options.SetDefault("kessel.url", fmt.Sprintf("%s:%d", e.Hostname, e.Port))
				}
			}
		*/
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
