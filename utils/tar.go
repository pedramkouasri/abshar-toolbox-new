package utils

import (
	"os"
	"os/exec"
)

func ExtractTarFile(serviceName string, dir string) error {
	cmd := exec.Command("tar", "-zxf", "./temp/"+serviceName+".tar.gz", "-C", dir)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
