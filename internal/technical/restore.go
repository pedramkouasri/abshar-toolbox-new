package technical

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

func NewRestore(cnf config.Config, branch string, loading contracts.Loader) *technical {
	return &technical{
		dir:         path.Join(cnf.DockerComposeDir, "services/technical-risk-micro-service"),
		branch:      branch,
		serviceName: "technical",
		env:         utils.LoadEnv(path.Join(cnf.DockerComposeDir, "services/technical-risk-micro-service")),
		percent:     0,
		loading:     loading,
		cnf:         cnf,
	}
}

func (t *technical) Restore(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := t.runRestore(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Restore %s Completed", t.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Restore Package %s is failed: %v", t.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", t.serviceName))
		return ctx.Err()
	}
}

func (t *technical) runRestore(ctx context.Context) error {
	var err error

	err = t.exec(ctx, 10, "Clean Storage", func() error {
		commands := []string{"sh", "-c", `find storage/app/ -type f -links 2 -exec rm -f {} +`}
		cmd := exec.Command(commands[0], commands[1:]...)
		cmd.Dir = t.dir
		bufE := bytes.NewBuffer([]byte{})
		cmd.Stderr = bufE
		if out, err := cmd.Output(); err != nil {
			return fmt.Errorf("Cannot Remove File In Storage :%v err: %s out: %s", err, bufE.String(), out)
		}

		commands = []string{"sh", "-c", "find storage/app/ -type d ! -name 'app' -exec rm -rf {} +"}

		cmd = exec.Command(commands[0], commands[1:]...)
		cmd.Dir = t.dir
		cmd.Stderr = bufE
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Cannot Remove Folder In Storage :%v err: %s out: %s", err, bufE.String(), out)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Technical Clean Storage Failed Error is: %s", err)
	}

	err = t.exec(ctx, 40, "Technical Extracted Tar File", func() error {
		return utils.ExtractTarFile(t.serviceName, t.dir)
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 45, "Technical Move Sql File", func() error {
		cd, _ := os.Getwd()
		backupSqlDir := cd + "/backupSql"

		return os.Rename(fmt.Sprintf("%s/%s.sql", t.dir, t.serviceName), fmt.Sprintf("%s/%s.sql", backupSqlDir, t.serviceName))
	})
	if err != nil {
		return fmt.Errorf("Extract Tar File Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 60, "Technical Restore DB ", func() error {
		return utils.RestoreDatabase("technical", t.cnf.DockerComposeDir, t.env)
	})
	if err != nil {
		return fmt.Errorf("Technical Restore DB Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 70, "Technical Restore Code ", func() error {
		return utils.RestoreCode(t.dir)
	})
	if err != nil {
		return fmt.Errorf("Technical Restore Code Failed Error Is: %s", err)
	}

	err = t.exec(ctx, 100, "Technical Restore Branch", func() error {
		return utils.SwitchBranch(t.dir, t.branch)
	})
	if err != nil {
		return fmt.Errorf("Technical Restore Branch Failed Error is: %s", err)
	}

	return nil
}
