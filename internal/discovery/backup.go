package discovery

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

func NewBackup(cnf config.Config, branch string, loading contracts.Loader) *discovery {
	return &discovery{
		dir:         path.Join(cnf.DockerComposeDir, "services/asset-discovery"),
		branch:      branch,
		serviceName: "discovery",
		percent:     0,
		loading:     loading,
		cnf:         cnf,
	}
}

func (d *discovery) Backup(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := d.runBackup(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Backup %s Completed", d.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Backup Package %s is failed: %v", d.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", d.serviceName))
		return ctx.Err()
	}
}

func (d *discovery) runBackup(ctx context.Context) error {
	var err error

	err = d.exec(ctx, 10, "Discovery Create Branch", func() error {
		return utils.CreateBranch(d.dir, d.branch)
	})
	if err != nil {
		return fmt.Errorf("Discovery Create Branch Failed Error is: %s", err)
	}

	err = d.exec(ctx, 30, "Discovery Commit files", func() error {
		return d.commitIfChanges(ctx)
	})
	if err != nil {
		return fmt.Errorf("Discovery Commit Failed Error is: %s", err)
	}

	pwd, _ := os.Getwd()
	err = d.exec(ctx, 90, "Create Tar File", func() error {
		outputFile := pwd + "/temp/builds/" + d.serviceName + ".tar"

		tarCommands := []string{
			"nice",
			"--10",
			"tar",
			"-cf",
			outputFile,
		}

		cmd := exec.Command("sh", "-c", strings.Join(tarCommands, " "))
		cmd.Dir = d.dir

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

	err = d.exec(ctx, 100, "Discovery Gzip Complete", func() error {
		logger.Info(strings.Join([]string{"gzip", "-f", fmt.Sprintf("%s/%s.tar", pwd+"/temp/builds", d.serviceName)}, " "))
		cmd := exec.Command("gzip", "-f", fmt.Sprintf("%s/%s.tar", pwd+"/temp/builds", d.serviceName))
		cmd.Stderr = os.Stderr
		_, err := cmd.Output()
		return err
	})
	if err != nil {
		return fmt.Errorf("Discovery Gzip Failed Error Is: %s", err)
	}

	return nil
}

func (d *discovery) commitIfChanges(ctx context.Context) error {
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = d.dir
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Git diff Failed Error is : %v\n", err)
	}

	if stdOut.Len() > 0 {

		if err := utils.GitAdd(d.dir); err != nil {
			return fmt.Errorf("Git Add Failed Error is: %s", err)
		}

		cmd := exec.Command("git", "commit", "-m", fmt.Sprintf("backup %s", d.branch))
		cmd.Dir = d.dir

		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Git Commit Backup is Failed Err is: %v", err)
		}
	}

	return nil
}
