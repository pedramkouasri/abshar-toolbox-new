package discovery

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewUpdate(cnf config.Config, version string, loading contracts.Loader) *discovery {
	return &discovery{
		dir:           path.Join(cnf.DockerComposeDir, "services/asset-discovery"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		tag2:          version,
		percent:       0,
		loading:       loading,
		cnf:           cnf,
		serviceName:   "discovery",
		containerName: "assetDiscovery",
	}
}

func (d *discovery) Update(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := d.runUpdate(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Update %s Completed", d.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Update Package %s is failed: %v", d.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", d.serviceName))
		return ctx.Err()
	}
}

func (d *discovery) runUpdate(ctx context.Context) error {
	var err error

	//run when service not exists
	if !utils.DirectoryExists(d.dir) {
		d.setPercent(100)
		logger.Info("Service Asset Discovery Not Exists")
		return nil
	}

	err = d.exec(ctx, 30, "Discovery Backup File Complete With git", func() error {
		return utils.BackupFileWithGit(d.dir, d.branch)
	})
	if err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}

	err = d.exec(ctx, 70, "Discovery Extracted Tar File", func() error {
		return utils.ExtractTarFile(d.serviceName, d.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = d.exec(ctx, 100, "Discovery Restart Service", func() error {
		return utils.RestartService(d.containerName, d.cnf.DockerComposeDir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Restart Service Discovery : %s", err)
	}

	return nil
}
