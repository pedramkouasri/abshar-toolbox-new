package baadbaan

import (
	"bytes"
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
		if err := b.runRestore(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Restore %s Completed", b.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Restore Package %s is failed: %v", b.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", b.serviceName))
		return ctx.Err()
	}
}

func (b *baadbaan) runRestore(ctx context.Context) error {
	var err error

	err = b.exec(ctx, 10, "Clean Storage", func() error {
		commands := []string{"find", "storage/app/", "-type", "f", "-links", "2", "-exec", "rm", "-f", "{}", ";"}
		cmd := exec.Command(commands[0], commands[1:]...)
		cmd.Dir = b.dir
		bufE := bytes.NewBuffer([]byte{})
		cmd.Stderr = bufE
		if out, err := cmd.Output(); err != nil {
			return fmt.Errorf("Cannot Remove File In Storage :%v err: %s out: %s", err, bufE.String(), out)
		}

		commands = []string{"find", "storage/app/", "-type", "d", "!", "-name", "'patches'", "!", "-name", "'versions'", "!", "-name", "'backup'", "-exec", "rm", "-rf", "{}", ";"}

		// cmd = exec.Command(commands[0], commands[1:]...)
		// cmd.Dir = b.dir
		// cmd.Stderr = bufE
		// if out, err := cmd.Output(); err != nil {
		// 	return fmt.Errorf("Cannot Remove Folder In Storage :%v err: %s out: %s", err, bufE.String(), out)
		// }
		return nil
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Clean Storage Failed Error is: %s", err)
	}
	return nil

	err = b.exec(ctx, 40, "Baadbaan Extracted Tar File", func() error {
		return utils.ExtractTarFile(b.serviceName, b.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 45, "Baadbaan Move Sql File", func() error {
		cd, _ := os.Getwd()
		backupSqlDir := cd + "/backupSql"

		return os.Rename(fmt.Sprintf("%s/%s.sql", b.dir, b.serviceName), fmt.Sprintf("%s/%s.sql", backupSqlDir, b.serviceName))
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 60, "Baadbaan Restore DB ", func() error {
		return utils.RestoreDatabase("baadbaan", b.cnf.DockerComposeDir, b.env)
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Restore DB Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 70, "Baadbaan Restore Code ", func() error {
		return utils.RestoreCode(b.dir)
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Restore Code Failed Error Is: %s", err)
	}

	err = b.exec(ctx, 100, "Baadbaan Restore Branch", func() error {
		return utils.SwitchBranch(b.dir, b.branch)
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Restore Branch Failed Error is: %s", err)
	}

	return nil
}
