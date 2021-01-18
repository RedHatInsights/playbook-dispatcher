package utils

import (
	"io"
	"os"
	"playbook-dispatcher/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/RedHatInsights/cloudwatch"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/spf13/viper"
)

var sugar *zap.SugaredLogger

func GetLoggerOrDie() *zap.SugaredLogger {
	if sugar == nil {
		cfg := config.Get()

		logCfg := zap.NewProductionConfig()
		logCfg.Level.UnmarshalText([]byte(cfg.GetString("log.level")))

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
	cloudWatchClient := cloudwatchlogs.New(cloudWatchSession)

	group := cloudwatch.NewGroup(cfg.GetString("log.cw.group"), cloudWatchClient)
	w, err := group.Create(hostname)

	if err != nil {
		return nil, err
	}

	cwc := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		wrapWriter(w),
		zap.NewAtomicLevelAt(level.Level()),
	)

	cloudwatch := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, cwc)
	})

	return cloudwatch, nil
}

func wrapWriter(w io.Writer) zapcore.WriteSyncer {
	switch w := w.(type) {
	case *cloudwatch.Writer:
		return &writerWrapper{w}
	default:
		return zapcore.AddSync(w)
	}
}

type writerWrapper struct {
	*cloudwatch.Writer
}

// this method is required by zapcore
func (w writerWrapper) Sync() error {
	return w.Flush()
}
