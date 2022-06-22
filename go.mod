module playbook-dispatcher

go 1.16

require (
	github.com/RedHatInsights/tenant-utils v1.0.0
	github.com/atombender/go-jsonschema v0.9.1-0.20220615213626-45dcfa31db10 // indirect
	github.com/aws/aws-sdk-go v1.38.51
	github.com/confluentinc/confluent-kafka-go v1.5.2
	github.com/deepmap/oapi-codegen v1.4.2
	github.com/getkin/kin-openapi v0.36.0
	github.com/ghodss/yaml v1.0.0
	github.com/globocom/echo-prometheus v0.1.2
	github.com/golang-migrate/migrate/v4 v4.15.1
	github.com/google/uuid v1.3.0
	github.com/kulshekhar/fungen v0.0.0-20180224173454-7ed9e430524b // indirect
	github.com/labstack/echo/v4 v4.1.17
	github.com/mec07/cloudwatchwriter v0.2.4
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/qri-io/jsonschema v0.2.0
	github.com/redhatinsights/app-common-go v1.2.0
	github.com/redhatinsights/platform-go-middlewares v0.17.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac
	gorm.io/driver/postgres v1.0.8
	gorm.io/gorm v1.21.4
)

replace gopkg.in/oleiade/lane.v1 v1.0.0 => github.com/oleiade/lane v1.0.1
