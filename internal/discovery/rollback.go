package discovery

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewRollback(cnf config.Config, version string, percent int) *discovery {
	return &discovery{
		dir:           path.Join(cnf.DockerComposeDir, "services/asset-discovery"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		serviceName:   "discovery",
		containerName: "assetDiscovery",
		percent:       percent,
		tag2:          version,
		cnf:           cnf,
	}
}

func (d *discovery) Rollback(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := d.runRollback(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Rollback %s Completed", d.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Rollback Package %s is failed: %v", d.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Rollback Canceled", d.serviceName))
		return ctx.Err()
	}
}

func (d *discovery) runRollback(ctx context.Context) error {
	//run when service not exists
	if !utils.DirectoryExists(d.dir) {
		return nil
	}

	if err := utils.RestoreCode(d.dir); err != nil {
		return fmt.Errorf("Discovery Restore Code Failed %v ", err)
	}

	logger.Info("Discovery Restore Code")

	if d.percent == 100 {
		if err := utils.RestartService(d.containerName, d.cnf.DockerComposeDir); err != nil {
			return fmt.Errorf("Cannot Discovery Restart Service In Rollback : %s", err)
		}
		logger.Info("Discovery Restart Service In Rollback")
	}

	return nil
}
