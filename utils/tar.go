package utils

import (
	"encoding/json"
	"fmt"
	"log"
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

func CreateTarFile(dir string, serviceName string) error {
	// tar -cf patch.tar --files-from=diff.txt
	cmd := exec.Command("tar", "-cf", "./patch.tar", fmt.Sprintf("--files-from=/temp/%s/diff.txt", serviceName))

	cmd.Dir = dir

	if _, err := cmd.Output(); err != nil {
		if err.Error() != "exit status 2" {
			return err
		}
	}
	return nil
}

func AddDiffPackageToTarFile(dir string, serviceName string) error {
	diffFile, err := getDiffPackages(serviceName)
	if err != nil {
		return err
	}
	for packageName := range diffFile {
		cmd := exec.Command("tar", "-rf", "./patch.tar", "vendor/"+packageName)
		cmd.Dir = dir
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func getDiffPackages(serviceName string) (map[string][]string, error) {
	//TODO::remove
	file, err := os.Open(fmt.Sprintf("/temp/%s/composer-lock-diff.json", serviceName))
	if err != nil {
		return nil, fmt.Errorf("cannot open composer-lock-diff %v", err)
	}
	defer file.Close()

	type ChangesType struct {
		Changes map[string][]string `json:"changes"`
	}

	changesInstance := ChangesType{}

	if err := json.NewDecoder(file).Decode(&changesInstance); err != nil {
		log.Fatal(err)
	}

	for index, packageName := range changesInstance.Changes {
		if packageName[1] == "REMOVED" {
			delete(changesInstance.Changes, index)
		}
	}

	return changesInstance.Changes, nil
}

func GzipTarFile(serviceName string) error {
	// cd {baadbaan_path} && gzip -f patch.tar
	cmd := exec.Command("gzip", "-f", fmt.Sprintf("/temp/%s/patch.tar", serviceName))

	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}
