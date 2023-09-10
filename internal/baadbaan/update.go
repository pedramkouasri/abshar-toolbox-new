package baadbaan

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewUpdate(cnf config.Config, version string, loading contracts.Loader) *baadbaan {
	return &baadbaan{
		dir:           path.Join(cnf.DockerComposeDir, "baadbaan_new"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		serviceName:   "baadbaan",
		containerName: "baadbaan_php",
		env:           utils.LoadEnv(path.Join(cnf.DockerComposeDir, "baadbaan_new")),
		percent:       0,
		loading:       loading,
	}
}

func (b *baadbaan) Update(ctx context.Context) error {

	completeSignal := make(chan bool)
	go func() {
		defer close(completeSignal)
		if err := b.runUpdate(ctx); err != nil {
			completeSignal <- false
		}
	}()

	select {
	case res, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Update %s Completed", b.serviceName))
			return nil
		}

		if res {
			return nil
		}

		return fmt.Errorf("Service Update %s is failed", b.serviceName)

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", b.serviceName))
		return ctx.Err()
	}
}

func (b *baadbaan) runUpdate(ctx context.Context) error {
	var err error

	err = b.exec(ctx, 10, "Baadbaan Changed Permission", func() error {
		return utils.ChangePermision("www-data", b.dir)
	})
	if err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}

	err = b.exec(ctx, 30, "Baadbaan Backup File Complete With git", func() error {
		return utils.BackupFileWithGit(b.dir, b.branch)
	})
	if err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 40, "Baadbaan Backup Database Complete", func() error {
		return utils.BackupDatabase(b.branch, b.env)
	})
	if err != nil {
		return fmt.Errorf("Backup Database Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 45, "Baadbaan Config Clear Complete", func() error {
		return utils.ConfigClear(b.dir)
	})
	if err != nil {
		return fmt.Errorf("Config Clear Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 50, "Baadbaan Extracted Tar File", func() error {
		return utils.ExtractTarFile(b.serviceName, b.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 65, "Baadbaan Composer Dump Autoload complete", func() error {
		return utils.ComposerDumpAutoload(b.containerName)
	})
	if err != nil {
		return fmt.Errorf("Composer Dump Autoload Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 90, "Baadbaan Migrated Database", func() error {
		return utils.MigrateDB(b.containerName)
	})
	if err != nil {
		return fmt.Errorf("Migrate Database Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 95, "Baadbaan View Cleared", func() error {
		return utils.ViewClear(b.containerName)
	})
	if err != nil {
		return fmt.Errorf("View Clear Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 97, "Baadbaan Config Cleared Completed", func() error {
		return utils.ConfigClear(b.dir)
	})
	if err != nil {
		return fmt.Errorf("Config Clear Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 98, "Baadbaan Config Cache Completed", func() error {
		return utils.ConfigCache(b.containerName)
	})
	if err != nil {
		return fmt.Errorf("Config Cache Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 100, "Baadbaan Changed Permission", func() error {
		return utils.ChangePermision("www-data", b.dir)
	})
	if err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}

	return nil
}
