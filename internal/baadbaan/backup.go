package baadbaan

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

func NewBackup(cnf config.Config, branch string, loading contracts.Loader) *baadbaan {
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

func (b *baadbaan) Backup(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := b.runBackup(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Backup %s Completed", b.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Backup Package %s is failed: %v", b.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", b.serviceName))
		return ctx.Err()
	}
}

func (b *baadbaan) runBackup(ctx context.Context) error {
	var err error

	err = b.exec(ctx, 10, "Baadbaan Create Branch", func() error {
		return utils.CreateBranch(b.dir, b.branch)
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Create Branch Failed Error is: %s", err)
	}

	err = b.exec(ctx, 30, "Baadbaan Commit files", func() error {
		return b.commitIfChanges(ctx)
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Commit Failed Error is: %s", err)
	}

	err = b.exec(ctx, 40, "Baadbaan Backup Database Complete", func() error {
		return utils.BackupDatabase(b.serviceName, b.cnf.DockerComposeDir, b.env)
	})
	if err != nil {
		return fmt.Errorf("Backup Database Failed Error Is: %s", err)
	}

	pwd, _ := os.Getwd()
	err = b.exec(ctx, 90, "Create Tar File", func() error {
		backupSqlDir := pwd + "/backupSql"

		dbPath := backupSqlDir + "/" + b.serviceName + ".sql"

		pathes := []string{
			"storage/app",
			dbPath,
		}

		outputFile := pwd + "/temp/builds/" + b.serviceName + ".tar"

		excludes := []string{
			"**/patches",
			"**/versions",
			"**/backup",
			"support.txt",
			"exp.txt",
		}

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
		cmd.Dir = b.dir

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

	err = b.exec(ctx, 100, "Baadbaan Gzip Complete", func() error {
		cmd := exec.Command("gzip", "-f", fmt.Sprintf("%s/%s.tar", pwd+"/temp/builds", b.serviceName))
		cmd.Stderr = os.Stderr
		_, err := cmd.Output()
		return err
	})
	if err != nil {
		return fmt.Errorf("Baadbaan Gzip Failed Error Is: %s", err)
	}

	return nil
}

func (b *baadbaan) commitIfChanges(ctx context.Context) error {
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = b.dir
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Git diff Failed Error is : %v\n", err)
	}

	if stdOut.Len() > 0 {

		if err := utils.GitAdd(b.dir); err != nil {
			return fmt.Errorf("Git Add Failed Error is: %s", err)
		}

		cmd := exec.Command("git", "commit", "-m", fmt.Sprintf("backup %s", b.branch))
		cmd.Dir = b.dir

		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Git Commit Backup is Failed Err is: %v", err)
		}
	}

	return nil
}
