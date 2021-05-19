package api

import (
	"context"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/common/utils/test"
	"sync"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"

	"github.com/spf13/viper"
)

func WithApi(cfg *viper.Viper) {
	var cancel context.CancelFunc

	BeforeSuite(func() {
		var ctx context.Context
		ctx, cancel = context.WithCancel(test.TestContext())
		Start(ctx, cfg, make(chan error, 1), &utils.ProbeHandler{}, &utils.ProbeHandler{}, &sync.WaitGroup{}, echo.New())
	})

	AfterSuite(func() {
		if cancel != nil {
			cancel()
		}
	})
}
