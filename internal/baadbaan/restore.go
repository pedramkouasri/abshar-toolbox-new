package baadbaan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewRestore(cnf config.Config, branch string, loading contracts.Loader) *baadbaan {
	return &baadbaan{
		tempDir:     cnf.TempDir + "/baadbaan",
		dir:         path.Join(cnf.DockerComposeDir, "baadbaan_new"),
		branch:      branch,
		serviceName: "baadbaan",
		env:         utils.LoadEnv(path.Join(cnf.DockerComposeDir, "baadbaan_new")),
		percent:     0,
		loading:     loading,
		cnf:         cnf,
	}
}

func (b *baadbaan) Restore(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := b.runUpdate(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Update %s Completed", b.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Update Package %s is failed: %v", b.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", b.serviceName))
		return ctx.Err()
	}
}

func (b *baadbaan) runRestore(ctx context.Context) error {
	var err error

	if err := utils.RestoreDatabase("baadbaan", b.cnf.DockerComposeDir, b.env); err != nil {
		return fmt.Errorf("Baadbaan Restore DB Failed %v ", err)
	}

	logger.Info("Baadbaan Restore DB")

	if err := utils.RestoreCode(b.dir); err != nil {
		return fmt.Errorf("Baadbaan Restore Code Failed %v ", err)
	}

	logger.Info("Baadbaan Restore Code")

	err = b.exec(ctx, 10, "Baadbaan Restore Branch", func() error {
		return utils.SwitchBranch(b.dir, b.branch)
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Restore Branch Failed Error is: %s", err)
	}

	logger.Info("Baadbaan Restored Branch")

	err = b.exec(ctx, 10, "Clean Storage", func() error {
		cmd := exec.Command("rm", "-r", "find storage/app/ -type f -links 2")
		cmd.Dir = b.dir
		cmd.Stderr = os.Stderr
		if _, err := cmd.Output(); err != nil {
			return fmt.Errorf("Cannot Remove File In Storage :%v", err)
		}

		cmd = exec.Command("rm", "-r", "find storage/app/ -type d ! -name 'patches' ! -name 'versions' ! -name 'backup'")
		cmd.Dir = b.dir
		cmd.Stderr = os.Stderr
		if _, err := cmd.Output(); err != nil {
			return fmt.Errorf("Cannot Remove Folder In Storage :%v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Clean Storage Failed Error is: %s", err)
	}

	// err = b.exec(ctx, 100, "Restore Storage", func() error {
	// 	return nil
	// })
	// if err != nil {
	// 	return fmt.Errorf("Baadbaan Restore Storage Failed Error is: %s", err)
	// }

	return nil
}
