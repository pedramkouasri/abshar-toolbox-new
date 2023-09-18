package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func BackupFileWithGit(dir string, branch string) error {
	var output []byte
	stdOut := bytes.NewBuffer(output)

	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = dir
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr

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

	cmd = exec.Command("git", "commit", "-m", fmt.Sprintf("backup befor update patch %s time: %d", branch, time.Now().Unix()))
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git Commit Backup is Failed Err is: %v", err)
	}

	return nil
}

func RestoreCode(dir string) error {
	if err := AddSafeDirectory(dir); err != nil {
		return fmt.Errorf("Cannot git Safe Direectory :%v", err)
	}

	cmd := exec.Command("git", "reset", "--hard")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("Cannot git reset :%v", err)
	}

	cmd = exec.Command("git", "clean", "-fd")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("Cannot git clean :%v", err)
	}

	return nil
}

func RemoveTag(dir string, branch string) {
	cmd := exec.Command("git", "tag", "-d", branch)
	cmd.Dir = dir
	cmd.Output()
}

func Fetch(dir string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("git --git-dir %s/.git  fetch", dir))

	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func GetDiff(dir string, tag1 string, tag2 string, excludePath []string, appendPatch []string, serviceName string) error {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter", "ACMR", tag1, tag2)

	bufE := bytes.NewBuffer([]byte{})
	cmd.Stderr = bufE
	cmd.Dir = dir

	res, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("%v %s", err, bufE.String())
	}

	if len(excludePath) > 0 {
		s := string(res)
		for _, path := range excludePath {
			s = strings.ReplaceAll(string(res), path, "")
		}
		res = []byte(s)
	}

	if len(appendPatch) > 0 {
		s := string(res)
		for _, path := range appendPatch {
			s = fmt.Sprintf("%s\n%s", s, path)
		}
		res = []byte(s)
	}

	currentDirectory, _ = os.Getwd()
	outPath := currentDirectory + "/temp/" + serviceName + "/diff.txt"
	// if FileExists(outPath) == false {
	// 	if _, err := os.Create(outPath); err != nil {
	// 		return fmt.Errorf("can not create %s file %v", outPath, err)
	// 	}
	// }

	// return ioutil.WriteFile(outPath, res, 0666)

	return os.WriteFile(outPath, res, 0666)
}

func SwitchBranch(dir string, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func AddSafeDirectory(dir string) error {
	command := strings.Split(gitSafeDirectory, " ")
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
