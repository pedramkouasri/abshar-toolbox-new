package docker

import (
	"context"
	"fmt"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/suppervisor"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewUpdate(cnf config.Config, version string, loading contracts.Loader) *docker {
	return &docker{
		dir:         cnf.DockerComposeDir,
		branch:      fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		tag2:        version,
		percent:     0,
		loading:     loading,
		cnf:         cnf,
		serviceName: "docker",
	}
}

func (d *docker) Update(ctx context.Context) error {

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

func (d *docker) runUpdate(ctx context.Context) error {
	var err error

	err = d.exec(ctx, 30, "Docker Backup File Complete With git", func() error {
		return utils.BackupFileWithGit(d.dir, d.branch)
	})
	if err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}

	err = d.exec(ctx, 80, "Docker Extracted Tar File", func() error {
		return utils.ExtractTarFile(d.serviceName, d.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = d.exec(ctx, 100, "Updated Config Supervisor", func() error {
		return suppervisor.ReloadConfig()
	})
	if err != nil {
		return fmt.Errorf("Cannot Update Config Supervisor : %s", err)
	}

	return nil
}
