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
	}
}

func (b *baadbaan) Rollback(ctx context.Context) error {

	completeSignal := make(chan bool)
	go func() {
		defer close(completeSignal)
		if err := b.runRollback(ctx); err != nil {
			completeSignal <- false
		}
	}()

	select {
	case res, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Rollback %s Completed", b.serviceName))
			return nil
		}

		if res {
			return nil
		}

		return fmt.Errorf("Service Rollback %s is failed", b.serviceName)

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

	if err := utils.RestoreDatabase(b.branch, b.env); err != nil {
		return fmt.Errorf("Baadbaan Restore DB Failed %v ", err)
	}

	logger.Info("Baadbaan Restore DB")

	if err := utils.RestoreCode(b.dir); err != nil {
		return fmt.Errorf("Baadbaan Restore Code Failed %v ", err)
	}

	logger.Info("Baadbaan Restore Code")

	return nil
}
