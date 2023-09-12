package docker

import (
	"context"
	"fmt"
	"log"
	"os"

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

func NewGenerator(cnf config.Config, tag1 string, tag2 string, loading contracts.Loader) *docker {

	if tag1 == "" {
		log.Fatal("tag 1 not initialized")
	}

	if tag2 == "" {
		log.Fatal("tag 2 not initialized")
	}

	return &docker{
		tempDir:     cnf.TempDir + "/docker",
		outFile:     cnf.TempDir + "/builds/docker.tar.gz",
		dir:         cnf.DockerComposeDir,
		tag1:        tag1,
		tag2:        tag2,
		percent:     0,
		loading:     loading,
		serviceName: "docker",
	}
}

func (d *docker) GetFilePath() string {
	return d.outFile
}

func (d *docker) Generate(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := d.runGenerate(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Generate Package %s Completed", d.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Generate Package %s is failed: %v", d.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Generate Package Canceled", d.serviceName))
		return ctx.Err()
	}
}

func (d *docker) runGenerate(ctx context.Context) error {
	var err error

	err = os.Mkdir(d.tempDir, 0755)
	if err != nil {
		return err
	}

	d.exec(ctx, 5, "Removed Tag", func() error {
		utils.RemoveTag(d.dir, d.tag2)
		return nil
	})

	err = d.exec(ctx, 10, "Fetch Sync With Git Server", func() error {
		return utils.Fetch(d.dir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Fetch: %v", err)
	}

	err = d.exec(ctx, 20, "Get Diff Code", func() error {
		return utils.GetDiff(d.dir, d.tag1, d.tag2, excludePath, appendPath, d.serviceName)
	})
	if err != nil {
		return fmt.Errorf("Cannot Get Diff: %v", err)
	}

	err = d.exec(ctx, 30, "Create Tar File", func() error {
		return utils.CreateTarFile(d.dir, d.tempDir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Create Tar File: %v", err)
	}

	err = d.exec(ctx, 90, "Copy Tar File To Temp Directory", func() error {
		return os.Rename(d.dir+"/patch.tar", d.tempDir+"/patch.tar")
	})
	if err != nil {
		return fmt.Errorf("Cannot Copy Tar File To Temp Directory: %v", err)
	}

	err = d.exec(ctx, 98, "Gzip Tar File", func() error {
		return utils.GzipTarFile(d.tempDir)
	})
	if err != nil {
		return fmt.Errorf("Cannot Gzip Tar File: %v", err)
	}

	err = d.exec(ctx, 100, "Gzip Tar File", func() error {
		return os.Rename(d.tempDir+"/patch.tar.gz", d.GetFilePath())
	})
	if err != nil {
		return fmt.Errorf("Cannot Gzip Tar File: %v", err)
	}

	return nil
}
