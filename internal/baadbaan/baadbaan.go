package baadbaan

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type baadbaan struct {
	dir           string
	branch        string
	serviceName   string
	containerName string
	env           *utils.ConfigService
	percent       int
	loading       contracts.Loader
}

func NewBaadbaan(cnf config.Config, version string, loading contracts.Loader) *baadbaan {
	return &baadbaan{
		dir:           path.Join(cnf.DockerComposeDir, "baadbaan_new"),
		branch:        fmt.Sprintf("patch-before-update-%s-%d", version, cnf.GetStartTime()),
		serviceName:   "baadbaan",
		containerName: "baadbaan_new",
		env:           utils.LoadEnv(path.Join(cnf.DockerComposeDir, "baadbaan_new")),
		percent:       0,
		loading:       loading,
	}
}

func (b *baadbaan) Update(ctx context.Context) error {

	completeSignal := make(chan bool)
	go func() {
		defer close(completeSignal)

		if err := ctx.Err(); err != nil {
			completeSignal <- false
			return
		}

		b.setPercent(10)

	}()

	select {
	case res, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("%s Completed", b.serviceName))
			return nil
		}

		if res {
			return nil
		}

		return fmt.Errorf("Service %s is failed", b.serviceName)

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", b.serviceName))
		return ctx.Err()
	}

	if err := utils.ChangePermision("www-data", b.dir); err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}

	// progress(types.Process{
	// 	State:   10,
	// 	Message: "Changed Permission",
	// })

	if err := utils.BackupFileWithGit(b.dir, b.branch); err != nil {
		return fmt.Errorf("Backup File With GIt Failed Error Is: %s", err)
	}
	// progress(types.Process{
	// 	State:   30,
	// 	Message: "Backup File Complete With git",
	// })

	if err := utils.BackupDatabase(b.serviceName, b.env); err != nil {
		return fmt.Errorf("Backup Database Failed Error Is: %s", err)
	}
	// progress(types.Process{
	// 	State:   40,
	// 	Message: "Backup Database Complete",
	// })

	if err := utils.ConfigClear(b.dir); err != nil {
		return fmt.Errorf("Config Clear Failed Error Is: %s", err)
	}

	if err := utils.ExtractTarFile(b.serviceName, b.dir); err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}
	// progress(types.Process{
	// 	State:   50,
	// 	Message: "Extracted Tar File",
	// })

	if err := utils.ComposerDumpAutoload(b.containerName); err != nil {
		return fmt.Errorf("Composer Dump Autoload Failed Error Is: %s", err)
	}
	// progress(types.Process{
	// 	State:   65,
	// 	Message: "Composer Dump Autoload complete",
	// })

	if err := utils.MigrateDB(b.containerName); err != nil {
		return fmt.Errorf("Migrate Database Failed Error Is: %s", err)
	}

	// progress(types.Process{
	// 	State:   90,
	// 	Message: "Migrated Database",
	// })

	if err := utils.ViewClear(b.containerName); err != nil {
		return fmt.Errorf("View Clear Failed Error Is: %s", err)
	}

	// progress(types.Process{
	// 	State:   95,
	// 	Message: "View Cleared",
	// })

	if err := utils.ConfigClear(b.dir); err != nil {
		return fmt.Errorf("Config Clear Failed Error Is: %s", err)
	}

	// progress(types.Process{
	// 	State:   97,
	// 	Message: "Config Cache Completed",
	// })

	if err := utils.ConfigCache(b.containerName); err != nil {
		return fmt.Errorf("Config Cache Failed Error Is: %s", err)
	}

	// progress(types.Process{
	// 	State:   98,
	// 	Message: "Config Cache Completed",
	// })

	if err := utils.ChangePermision("www-data", b.dir); err != nil {
		return fmt.Errorf("Change Permission has Error : %s", err)
	}
	// progress(types.Process{
	// 	State:   100,
	// 	Message: "Changed Permission",
	// })

	return nil
}

func (b *baadbaan) Rollback() {
	fmt.Println("Start Roolbacking")
	time.Sleep(time.Second * 30)
	fmt.Println("End Roolbacking")
}

func (b *baadbaan) setPercent(percent int) {
	b.percent = percent
	b.loading.Update(b.serviceName, b.percent)
}
