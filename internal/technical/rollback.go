package technical

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewRollback(cnf config.Config, version string, percent int) *technical {
	return &technical{
		dir:           path.Join(cnf.DockerComposeDir, "services/technical-risk-micro-service"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		serviceName:   "technical",
		containerName: "technical_risk_php",
		env:           utils.LoadEnv(path.Join(cnf.DockerComposeDir, "services/technical-risk-micro-service")),
		percent:       percent,
		tag2:          version,
		cnf:           cnf,
	}
}

func (t *technical) Rollback(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := t.runRollback(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Rollback %s Completed", t.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Rollback Package %s is failed: %v", t.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Rollback Canceled", t.serviceName))
		return ctx.Err()
	}
}

func (t *technical) runRollback(ctx context.Context) error {
	//run when service not exists
	if !utils.DirectoryExists(t.dir) {
		return nil
	}

	logger.Info(fmt.Sprintf("%d", t.percent))
	if t.percent < 50 {
		return nil
	}

	if err := utils.RestoreDatabase(t.tag2, t.cnf.DockerComposeDir, t.env); err != nil {
		return fmt.Errorf("Technical Restore DB Failed %v ", err)
	}

	logger.Info("Technical Restore DB")

	if err := utils.RestoreCode(t.dir); err != nil {
		return fmt.Errorf("Technical Restore Code Failed %v ", err)
	}

	logger.Info("Technical Restore Code")

	if t.percent == 100 {
		if err := utils.RestartService(t.containerName, t.cnf.DockerComposeDir); err != nil {
			return fmt.Errorf("Cannot Technical Restart Service In Rollback : %s", err)
		}
		logger.Info("Technical Restart Service In Rollback")
	}

	return nil
}
