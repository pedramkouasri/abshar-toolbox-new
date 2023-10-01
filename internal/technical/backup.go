package technical

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewBackup(cnf config.Config, branch string, loading contracts.Loader) *technical {
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

func (t *technical) Backup(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := t.runBackup(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Backup %s Completed", t.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Backup Package %s is failed: %v", t.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", t.serviceName))
		return ctx.Err()
	}
}

func (t *technical) runBackup(ctx context.Context) error {
	var err error

	err = t.exec(ctx, 10, "Technical Create Branch", func() error {
		return utils.CreateBranch(t.dir, t.branch)
	})
	if err != nil {
		return fmt.Errorf("Technical Create Branch Failed Error is: %s", err)
	}

	err = t.exec(ctx, 30, "Technical Commit files", func() error {
		return t.commitIfChanges(ctx)
	})
	if err != nil {
		return fmt.Errorf("Technical Commit Failed Error is: %s", err)
	}

	err = t.exec(ctx, 40, "Technical Backup Database Complete", func() error {
		return utils.BackupDatabase(t.serviceName, t.cnf.DockerComposeDir, t.env)
	})
	if err != nil {
		return fmt.Errorf("Backup Database Failed Error Is: %s", err)
	}

	pwd, _ := os.Getwd()
	err = t.exec(ctx, 90, "Create Tar File", func() error {
		backupSqlDir := pwd + "/backupSql"

		dbPath := backupSqlDir + "/" + t.serviceName + ".sql"

		pathes := []string{
			"storage/app",
			dbPath,
		}

		outputFile := pwd + "/temp/builds/" + t.serviceName + ".tar"

		excludes := []string{}

		tarCommands := []string{
			"nice",
			"--10",
			"tar",
			"--transform",
			fmt.Sprintf("s,%s,,S", strings.Replace(backupSqlDir, "/", "", 1)+"/"),
			"-cf",
			outputFile,
		}

		for _, excludePath := range excludes {
			tarCommands = append(tarCommands, fmt.Sprintf("--exclude='%s'", excludePath))
		}

		for _, path := range pathes {
			tarCommands = append(tarCommands, path)
		}

		// cmd := exec.Command(tarCommands[0], tarCommands[1:]...)
		cmd := exec.Command("sh", "-c", strings.Join(tarCommands, " "))
		cmd.Dir = t.dir

		if _, err := cmd.Output(); err != nil {
			if err.Error() != "exit status 2" {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Cannot Create Tar File: %v", err)
	}

	err = t.exec(ctx, 100, "Technical Gzip Complete", func() error {
		logger.Info(strings.Join([]string{"gzip", "-f", fmt.Sprintf("%s/%s.tar", pwd+"/temp/builds", t.serviceName)}, " "))
		cmd := exec.Command("gzip", "-f", fmt.Sprintf("%s/%s.tar", pwd+"/temp/builds", t.serviceName))
		cmd.Stderr = os.Stderr
		_, err := cmd.Output()
		return err
	})
	if err != nil {
		return fmt.Errorf("Technical Gzip Failed Error Is: %s", err)
	}

	return nil
}

func (t *technical) commitIfChanges(ctx context.Context) error {
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = t.dir
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Git diff Failed Error is : %v\n", err)
	}

	if stdOut.Len() > 0 {

		if err := utils.GitAdd(t.dir); err != nil {
			return fmt.Errorf("Git Add Failed Error is: %s", err)
		}

		cmd := exec.Command("git", "commit", "-m", fmt.Sprintf("backup %s", t.branch))
		cmd.Dir = t.dir

		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Git Commit Backup is Failed Err is: %v", err)
		}
	}

	return nil
}
