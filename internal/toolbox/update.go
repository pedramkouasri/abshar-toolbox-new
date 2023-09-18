package toolbox

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewUpdate(cnf config.Config, version string, loading contracts.Loader) *toolbox {
	return &toolbox{
		dir:         path.Join(cnf.DockerComposeDir, "services/update-toolbox"),
		branch:      fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		tag2:        version,
		percent:     0,
		loading:     loading,
		cnf:         cnf,
		serviceName: "toolbox",
	}
}

func (t *toolbox) Update(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := t.runUpdate(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Update %s Completed", t.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Update Package %s is failed: %v", t.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", t.serviceName))
		return ctx.Err()
	}
}

func (t *toolbox) runUpdate(ctx context.Context) error {
	var err error

	err = t.exec(ctx, 30, "Toolbox Backup File Complete With git", func() error {
		return utils.BackupFileWithGit(t.dir, t.branch)
	})
	if err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 100, "Toolbox Extracted Tar File", func() error {
		return utils.ExtractTarFile(t.serviceName, t.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	// err = t.exec(ctx, 100, "Toolbox Down All Service", func() error {
	// 	return utils.DockerDown(t.cnf.DockerComposeDir)
	// })
	// if err != nil {
	// 	return fmt.Errorf("Cannot Toolbox Down All Service : %s", err)
	// }

	return nil
}
