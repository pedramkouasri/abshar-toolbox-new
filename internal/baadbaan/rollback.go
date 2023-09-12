package baadbaan

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewRollback(cnf config.Config, version string, percent int) *baadbaan {
	return &baadbaan{
		dir:           path.Join(cnf.DockerComposeDir, "baadbaan_new"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		serviceName:   "baadbaan",
		containerName: "baadbaan_php",
		env:           utils.LoadEnv(path.Join(cnf.DockerComposeDir, "baadbaan_new")),
		percent:       percent,
		tag2:          version,
		cnf:           cnf,
	}
}

func (b *baadbaan) Rollback(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := b.runRollback(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Rollback %s Completed", b.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Rollback Package %s is failed: %v", b.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Rollback Canceled", b.serviceName))
		return ctx.Err()
	}
}

func (b *baadbaan) runRollback(ctx context.Context) error {
	logger.Info(fmt.Sprintf("%d", b.percent))
	if b.percent < 50 {
		return nil
	}

	if err := utils.RestoreDatabase(b.tag2, b.cnf.DockerComposeDir, b.env); err != nil {
		return fmt.Errorf("Baadbaan Restore DB Failed %v ", err)
	}

	logger.Info("Baadbaan Restore DB")

	if err := utils.RestoreCode(b.dir); err != nil {
		return fmt.Errorf("Baadbaan Restore Code Failed %v ", err)
	}

	logger.Info("Baadbaan Restore Code")

	if b.percent == 100 {
		if err := utils.RestartService(b.containerName, b.cnf.DockerComposeDir); err != nil {
			return fmt.Errorf("Cannot Baadbaan Restart Service In Rollback : %s", err)
		}
		logger.Info("Baadbaan Restart Service In Rollback")
	}

	return nil
}
