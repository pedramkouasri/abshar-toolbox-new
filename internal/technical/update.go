package technical

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewUpdate(cnf config.Config, version string, loading contracts.Loader) *technical {
	return &technical{
		dir:           path.Join(cnf.DockerComposeDir, "services/technical-risk-micro-service"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		tag2:          version,
		serviceName:   "technical",
		containerName: "technical_risk_php",
		env:           utils.LoadEnv(path.Join(cnf.DockerComposeDir, "services/technical-risk-micro-service")),
		percent:       0,
		loading:       loading,
		cnf:           cnf,
	}
}

func (t *technical) Update(ctx context.Context) error {

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

func (t *technical) runUpdate(ctx context.Context) error {
	var err error

	err = t.exec(ctx, 10, "Technical Changed Permission", func() error {
		return utils.ChangePermision("www-data", t.dir)
	})
	if err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}

	err = t.exec(ctx, 30, "Technical Backup File Complete With git", func() error {
		return utils.BackupFileWithGit(t.dir, t.branch)
	})
	if err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 40, "Technical Backup Database Complete", func() error {
		return utils.BackupDatabase(t.tag2, t.cnf.DockerComposeDir, t.env)
	})
	if err != nil {
		return fmt.Errorf("Backup Database Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 45, "Technical Config Clear Complete", func() error {
		return utils.ConfigClear(t.dir)
	})
	if err != nil {
		return fmt.Errorf("Config Clear Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 50, "Technical Extracted Tar File", func() error {
		return utils.ExtractTarFile(t.serviceName, t.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 65, "Technical Composer Dump Autoload complete", func() error {
		return utils.ComposerDumpAutoload(t.containerName)
	})
	if err != nil {
		return fmt.Errorf("Composer Dump Autoload Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 90, "Technical Migrated Database", func() error {
		return utils.MigrateDB(t.containerName)
	})
	if err != nil {
		return fmt.Errorf("Migrate Database Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 95, "Technical View Cleared", func() error {
		return utils.ViewClear(t.containerName)
	})
	if err != nil {
		return fmt.Errorf("View Clear Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 97, "Technical Config Cleared Completed", func() error {
		return utils.ConfigClear(t.dir)
	})
	if err != nil {
		return fmt.Errorf("Config Clear Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 98, "Technical Config Cache Completed", func() error {
		return utils.ConfigCache(t.containerName)
	})
	if err != nil {
		return fmt.Errorf("Config Cache Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 99, "Technical Changed Permission", func() error {
		return utils.ChangePermision("www-data", t.dir)
	})
	if err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}

	err = t.exec(ctx, 100, "Technical Restart Service", func() error {
		return utils.RestartService(t.containerName, t.cnf.DockerComposeDir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Restart Service Technical : %s", err)
	}

	return nil
}
