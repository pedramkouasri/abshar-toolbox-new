package utils

import (
	"bufio"
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

func ComposerChangedOrPanic(tempDir string) bool {
	diffFile, err := os.Open(fmt.Sprintf("%s/diff.txt", tempDir))
	if err != nil {
		panic(fmt.Sprintf("can not open %s/diff.txt: %v", tempDir, err))
	}

	defer diffFile.Close()

	scanner := bufio.NewScanner(diffFile)

	var exists bool = false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "composer.lock" {
			exists = true
			break
		}
	}

	// Check for any errors during scanning
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return exists
}

func ComposerInstall(containerName string) error {
	command := getCommand(composerInstallCommand, containerName)
	cmd := exec.Command(command[0], command[1:]...)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func GenerateDiffJson(dir string, tempDir string, tag1, tag2 string) error {

	file, err := os.Create(fmt.Sprintf("%s/composer-lock-diff.json", tempDir))
	if err != nil {
		return fmt.Errorf("can not create composer-lock-diff.json :%v", err)
	}

	cmd := exec.Command("composer-lock-diff", "--from", tag1, "--to", tag2, "--json", "--pretty", "--only-prod")
	cmd.Stdout = file
	// cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
