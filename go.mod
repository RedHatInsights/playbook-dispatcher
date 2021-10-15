module playbook-dispatcher

go 1.14

require (
	github.com/atombender/go-jsonschema v0.8.1-0.20200630165108-d692b4118fc7 // indirect
	github.com/aws/aws-sdk-go v1.36.28
	github.com/confluentinc/confluent-kafka-go v1.5.2
	github.com/deepmap/oapi-codegen v1.4.2
	github.com/getkin/kin-openapi v0.36.0
	github.com/ghodss/yaml v1.0.0
	github.com/globocom/echo-prometheus v0.1.2
	github.com/golang-migrate/migrate v3.5.4+incompatible
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/uuid v1.1.5
	github.com/kulshekhar/fungen v0.0.0-20180224173454-7ed9e430524b // indirect
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/mec07/cloudwatchwriter v0.2.4
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.1.0
	github.com/qri-io/jsonschema v0.2.0
	github.com/redhatinsights/app-common-go v1.2.0
	github.com/redhatinsights/platform-go-middlewares v0.7.1-0.20201009171810-b73c54b47a2d
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.7.1
	go.uber.org/ratelimit v0.2.0
	go.uber.org/zap v1.10.0
	gorm.io/driver/postgres v1.0.6
	gorm.io/gorm v1.20.11
)

replace gopkg.in/oleiade/lane.v1 v1.0.0 => github.com/oleiade/lane v1.0.1
