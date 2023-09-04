package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ComposerDumpAutoload(containerName string) error {

	err := os.Setenv("HOME", "/tmp")
	if err != nil {
		return fmt.Errorf("Error setting environment variable : %v", err)
	}

	var command []string
	command = getCommand(composerDumpCommand, containerName)

	cmd := exec.Command(command[0], command[1:]...)

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func MigrateDB(containerName string) error {
	var command []string = getCommand(migrateCommand, containerName)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func ViewClear(containerName string) error {
	var command []string = getCommand(viewClearCommand, containerName)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func ConfigCache(containerName string) error {
	var command []string = getCommand(configCacheCommand, containerName)

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func ConfigClear(dir string) error {
	err := filepath.Walk(dir+"/bootstrap/cache", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		split := strings.Split(path, "/")
		if strings.HasSuffix(split[len(split)-1], ".php") {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("remove file error %s", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk to filepath error in err  %s\n", err)
	}
	return nil
}
