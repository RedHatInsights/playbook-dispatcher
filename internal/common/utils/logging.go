package utils

import (
	"context"
	"os"
	"playbook-dispatcher/internal/common/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/labstack/echo/v4"
	"github.com/mec07/cloudwatchwriter"
	"github.com/spf13/viper"
)

var (
	sugar  *zap.SugaredLogger
	writer *cloudwatchwriter.CloudWatchWriter
)

func GetLoggerOrDie() *zap.SugaredLogger {
	if sugar == nil {
		cfg := config.Get()

		logCfg := zap.NewProductionConfig()
		err := logCfg.Level.UnmarshalText([]byte(cfg.GetString("log.level")))

		if err != nil {
			DieOnError(err)
		}

		options := []zap.Option{}

		if len(cfg.GetString("log.cw.accessKeyId")) > 0 {
			cwc, err := createCloudwatch(cfg, logCfg.Level)
			DieOnError(err)
			options = append(options, cwc)
		}

		log, err := logCfg.Build(options...)
		DieOnError(err)

		sugar = log.Sugar()
	}

	return sugar
}

func LogWithRequestId(log *zap.SugaredLogger, value string) *zap.SugaredLogger {
	return log.With("request_id", value)
}

func CloseLogger() {
	if writer != nil {
		writer.Close()
	}
}

func createCloudwatch(cfg *viper.Viper, level zap.AtomicLevel) (zap.Option, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	awsConf := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(
			cfg.GetString("log.cw.accessKeyId"), cfg.GetString("log.cw.secretAccessKey"), "",
		)).
		WithRegion(cfg.GetString("log.cw.region"))

	cloudWatchSession := session.Must(session.NewSession(awsConf))

	writer, err = cloudwatchwriter.New(cloudWatchSession, cfg.GetString("log.cw.group"), hostname)
	if err != nil {
		return nil, err
	}

	cwc := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		zap.NewAtomicLevelAt(level.Level()),
	)

	cloudwatch := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, cwc)
	})

	return cloudwatch, nil
}

type loggerKeyType int

const loggerKey loggerKeyType = iota

func SetLog(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

func GetLogFromContextIfAvailable(ctx context.Context) *zap.SugaredLogger {
	if value := ctx.Value(loggerKey); value == nil {
		return nil
	} else if log, ok := value.(*zap.SugaredLogger); ok {
		return log
	} else {
		panic("Unexpected logger type in context")
	}
}

func GetLogFromContext(ctx context.Context) *zap.SugaredLogger {
	log := GetLogFromContextIfAvailable(ctx)

	if log == nil {
		panic("Logger missing in context")
	}

	return log
}

func GetLogFromEcho(ctx echo.Context) *zap.SugaredLogger {
	return GetLogFromContext(ctx.Request().Context())
}

func WithRequestId(parent context.Context, requestId string) context.Context {
	return withKeyValue(parent, "request_id", requestId)
}

func WithCorrelationId(parent context.Context, correlationId string) context.Context {
	return withKeyValue(parent, "correlation_id", correlationId)
}

func WithAccount(parent context.Context, account string) context.Context {
	return withKeyValue(parent, "account", account)
}

func WithOrgId(parent context.Context, orgId string) context.Context {
	return withKeyValue(parent, "org_id", orgId)
}

func WithRequestType(parent context.Context, requestType string) context.Context {
	return withKeyValue(parent, "request_type", requestType)
}

func withKeyValue(parent context.Context, key, value string) context.Context {
	return SetLog(parent, GetLogFromContext(parent).With(key, value))
}
