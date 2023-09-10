package baadbaan

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

var excludePath = []string{
	".env",
	"vmanager.json",
}

var appendPath = []string{}

func NewGenerator(cnf config.Config, tag1 string, tag2 string, loading contracts.Loader) *baadbaan {

	if tag1 == "" {
		log.Fatal("tag 1 not initialized")
	}

	if tag2 == "" {
		log.Fatal("tag 2 not initialized")
	}

	return &baadbaan{
		dir:           path.Join(cnf.DockerComposeDir, "baadbaan_new"),
		tag1:          tag1,
		tag2:          tag2,
		serviceName:   "baadbaan",
		containerName: "baadbaan_new",
		env:           utils.LoadEnv(path.Join(cnf.DockerComposeDir, "baadbaan_new")),
		percent:       0,
	}
}

func (b *baadbaan) Generate(ctx context.Context) error {

	completeSignal := make(chan bool)
	go func() {
		defer close(completeSignal)
		if err := b.runGenerate(ctx); err != nil {
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

func (b *baadbaan) runGenerate(ctx context.Context) error {
	var err error

	b.exec(ctx, 5, "Removed Tag", func() error {
		utils.RemoveTag(b.dir, b.tag2)
		return nil
	})

	err = b.exec(ctx, 10, "Fetch Sync With Git Server", func() error {
		return utils.Fetch(b.dir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Fetch: %v", err)
	}

	err = b.exec(ctx, 20, "Get Diff Code", func() error {
		return utils.GetDiff(b.dir, b.tag1, b.tag2, excludePath, appendPath)
	})
	if err != nil {
		return fmt.Errorf("Cannot Get Diff: %v", err)
	}

	err = b.exec(ctx, 30, "Create Tar File", func() error {
		return utils.CreateTarFile(b.dir, b.serviceName)
	})
	if err != nil {
		return fmt.Errorf("Cannot Create Tar File: %v", err)
	}

	if utils.ComposerChangedOrPanic(b.serviceName) {

		err = b.exec(ctx, 50, "Switched Branch", func() error {
			return utils.SwitchBranch(b.dir, b.tag2)
		})
		if err != nil {
			return fmt.Errorf("Cannot Switch Branch: %v", err)
		}

		if err := utils.AddSafeDirectory(b.dir, b.containerName); err != nil {
			return fmt.Errorf("Cannot Add Safe Directory: %v", err)
		}

		err = b.exec(ctx, 60, "Composer Installed", func() error {
			return utils.ComposerInstall(b.containerName)
		})
		if err != nil {
			return fmt.Errorf("Cannot Composer Install: %v", err)
		}

		err = b.exec(ctx, 70, "Generate Json Diff Vendor", func() error {
			return utils.GenerateDiffJson(b.dir, b.serviceName, b.tag1, b.tag2)
		})
		if err != nil {
			return fmt.Errorf("Cannot Generate Json Diff Vendor: %v", err)
		}

		err = b.exec(ctx, 80, "Add Diff Package Vendor", func() error {
			return utils.AddDiffPackageToTarFile(b.dir, b.serviceName)
		})
		if err != nil {
			return fmt.Errorf("Cannot Add Diff Package Vendor: %v", err)
		}

	}

	err = b.exec(ctx, 90, "Copy Tar File To Temp Directory", func() error {
		return os.Rename(b.dir+"/patch.tar", "/temp/"+b.serviceName+"/patch.tar")
	})
	if err != nil {
		return fmt.Errorf("Cannot Copy Tar File To Temp Directory: %v", err)
	}

	err = b.exec(ctx, 100, "Gzip Tar File", func() error {
		return utils.GzipTarFile(b.serviceName)
	})
	if err != nil {
		return fmt.Errorf("Cannot Gzip Tar File: %v", err)
	}

	return nil
}
