package cmd

import (
	"context"
	"playbook-dispatcher/internal/common/config"
	"playbook-dispatcher/internal/common/db"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

func clean(cmd *cobra.Command, args []string) error {
	log := utils.GetLoggerOrDie()
	defer utils.CloseLogger()
	cfg := config.Get()
	ctx := utils.SetLog(context.Background(), log)

	db, sql := db.Connect(ctx, cfg)
	defer sql.Close()

	err := db.Transaction(func(tx *gorm.DB) error {
		log.Info("Cleaning up timed-out runs")

		var dbRuns []dbModel.Run

		result := tx.Model(&dbModel.Run{}).
			Where("runs.status", "running").
			Where("runs.created_at + runs.timeout * interval '1 second' <= NOW()").
			Select("id", "account", "correlation_id", "recipient").
			Find(&dbRuns)

		if result.Error != nil {
			return result.Error
		}

		if len(dbRuns) == 0 {
			log.Infow("No runs to update")
			return nil
		}

		ids := make([]string, len(dbRuns))
		for i, run := range dbRuns {
			log.Infow("Updating timed-out run", "run_id", run.ID.String(), "account", run.Account, "correlation_id", run.CorrelationID.String(), "recipient", run.Recipient.String())
			ids[i] = run.ID.String()
		}

		result = tx.Model(&dbModel.Run{}).
			Where("runs.id IN ?", ids).
			Update("status", "timeout")

		log.Infow("Finished updating timed-out runs", "rowCount", result.RowsAffected)

		if result.Error != nil {
			return result.Error
		}

		subQuery := tx.Model(&dbModel.RunHost{}).
			Select("run_hosts.id").
			Joins("INNER JOIN runs on runs.id = run_hosts.run_id").
			Where("runs.status", "timeout").
			Where("run_hosts.status", "running")

		result = tx.Model(&dbModel.RunHost{}).
			Where("run_hosts.id IN (?)", subQuery).
			Update("status", "timeout")

		log.Infow("Finished updating timed-out run_hosts", "rowCount", result.RowsAffected)

		return result.Error
	})

	if err != nil {
		log.Error(err)
	}

	return err
}
