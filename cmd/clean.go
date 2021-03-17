package cmd

import (
	"context"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/db"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/spf13/cobra"
)

func clean(cmd *cobra.Command, args []string) error {
	log := utils.GetLoggerOrDie()
	cfg := config.Get()
	ctx := utils.SetLog(context.Background(), log)

	db, sql := db.Connect(ctx, cfg)
	defer sql.Close()

	log.Info("Cleaning up timed-out runs")
	result := db.Model(&dbModel.Run{}).
		Where("runs.status='running'").
		Where("runs.created_at + runs.timeout * interval '1 second' <= NOW()").
		Update("status", "timeout")

	log.Infow("Finished updating timed-out runs", "rowCount", result.RowsAffected)

	return result.Error
}
