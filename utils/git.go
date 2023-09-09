package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TODO::remove
var current_time = time.Now()

func BackupFileWithGit(dir string, branch string) error {
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = dir
	cmd.Stdout = stdOut

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Git diff Failed Error is : %v\n", err)
	}

	if stdOut.Len() > 0 {
		if err := createBranch(dir, branch); err != nil {
			return fmt.Errorf("Create Branch Failed Error is : %v\n", err)
		}
		if err := gitAdd(dir); err != nil {
			return fmt.Errorf("Git Add Failed Error is : %v\n", err)
		}
		if err := gitCommit(dir, branch); err != nil {
			return fmt.Errorf("Git Commit Failed Error is : %v\n", err)
		}
	}

	return nil
}

func createBranch(dir string, branch string) error {
	cmd := exec.Command("git", strings.Fields(fmt.Sprintf("checkout -b %s", branch))...)

	cmd.Stdout = nil

	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func gitAdd(dir string) error {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if _, err := cmd.Output(); err != nil {
		return err
	}
	return nil
}

func gitCommit(dir string, branch string) error {

	err := os.Setenv("HOME", "/tmp")
	if err != nil {
		return fmt.Errorf("Error setting environment variable: %v", err)
	}

	cmd := exec.Command("git", "config", "--global", "user.email", "persianped@gmail.com")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git set config Email Failed Error Is: %v", err)
	}

	cmd = exec.Command("git", "config", "--global", "user.name", "pedram kousari")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git set config UserName Failed Error Is: %v", err)
	}

	cmd = exec.Command("git", "config", "--global", "--add", "safe.directory", dir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git Set Safe Directory Failed Error Is: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("backup befor update patch %s", branch, current_time.Unix()))
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git Commit Backup is Failed Err is: %v", err)
	}

	return nil
}

func RestoreCode(dir string) error {
	cmd := exec.Command("git", "reset", "--hard")
	cmd.Dir = dir
	if _, err := cmd.Output(); err != nil {
		return err
	}

	cmd = exec.Command("git", "clean", "-fd")
	cmd.Dir = dir
	if _, err := cmd.Output(); err != nil {
		return err
	}

	return nil
}
