package suppervisor

import (
	"fmt"
	"os"
	"os/exec"
)

func reloadConfig() error {
	command := []string{"supervisorctl", "update"}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func RestartAllService() error {
	if err := reloadConfig(); err != nil {
		return fmt.Errorf("cannot reload config %v", err)
	}

	return ReloadService("all")
}

func ReloadService(sericeName string) error {
	command := []string{"supervisorctl", "restart", sericeName}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
