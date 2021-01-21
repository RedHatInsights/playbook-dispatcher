package api

import (
	"context"
	"playbook-dispatcher/utils"

	. "github.com/onsi/ginkgo"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func WithApi(cfg *viper.Viper) {
	var stop func(ctx context.Context)

	BeforeSuite(func() {
		log := zap.NewNop().Sugar()
		//log, _ := zap.NewProduction()
		stop = Start(cfg, log, make(chan error, 1), &utils.ProbeHandler{Log: log}, &utils.ProbeHandler{Log: log})
	})

	AfterSuite(func() {
		if stop != nil {
			stop(context.TODO())
		}
	})
}
