module playbook-dispatcher

go 1.23.8

toolchain go1.24.2

require (
	github.com/RedHatInsights/tenant-utils v1.0.0
	github.com/aws/aws-sdk-go v1.49.6
	github.com/confluentinc/confluent-kafka-go/v2 v2.10.0
	github.com/getkin/kin-openapi v0.132.0
	github.com/ghodss/yaml v1.0.0
	github.com/globocom/echo-prometheus v0.1.2
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v4 v4.12.0
	github.com/mec07/cloudwatchwriter v0.2.6
	github.com/oapi-codegen/echo-middleware v1.0.2
	github.com/oapi-codegen/runtime v1.1.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/prometheus/client_golang v1.17.0
	github.com/qri-io/jsonschema v0.2.0
	github.com/redhatinsights/app-common-go v1.6.2
	github.com/redhatinsights/platform-go-middlewares v0.17.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.7.1
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8
	go.uber.org/zap v1.13.0
	golang.org/x/time v0.6.0
	gorm.io/driver/postgres v1.0.8
	gorm.io/gorm v1.21.4
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/jackc/pgx/v4 v4.18.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/qri-io/jsonpointer v0.1.1 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/oleiade/lane.v1 v1.0.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gopkg.in/oleiade/lane.v1 v1.0.0 => github.com/oleiade/lane v1.0.1

// See https://github.com/oapi-codegen/runtime/issues/50
replace github.com/chenzhuoyu/iasm v0.9.0 => github.com/cloudwego/iasm v0.2.0
