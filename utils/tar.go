package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func ExtractTarFile(serviceName string, dir string) error {
	cmd := exec.Command("tar", "-zxf", "./temp/"+serviceName+".tar.gz", "-C", dir)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func CreateTarFile(dir string, tempDir string) error {
	// tar -cf patch.tar --files-from=diff.txt
	cmd := exec.Command("tar", "-cf", "./patch.tar", fmt.Sprintf("--files-from=%s/diff.txt", tempDir))

	cmd.Dir = dir

	if _, err := cmd.Output(); err != nil {
		if err.Error() != "exit status 2" {
			return err
		}
	}
	return nil
}

func AddDiffPackageToTarFile(dir string, tempDir string) error {
	diffFile, err := getDiffPackages(tempDir)
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

func getDiffPackages(tempDir string) (map[string][]string, error) {
	//TODO::remove
	file, err := os.Open(fmt.Sprintf("%s/composer-lock-diff.json", tempDir))
	if err != nil {
		return nil, fmt.Errorf("cannot open composer-lock-diff %v", err)
	}
	defer file.Close()

	type ChangesType struct {
		Changes map[string][]string `json:"changes"`
	}

	changesInstance := ChangesType{}

	//Todo::remove
	// if err := json.NewDecoder(file).Decode(&changesInstance); err != nil {
	// 	return map[string][]string{}, fmt.Errorf("Cannot Decode composer-lock-diff :%v", err)
	// }

	for index, packageName := range changesInstance.Changes {
		if packageName[1] == "REMOVED" {
			delete(changesInstance.Changes, index)
		}
	}

	return changesInstance.Changes, nil
}

func GzipTarFile(tempDir string) error {
	// cd {baadbaan_path} && gzip -f patch.tar
	cmd := exec.Command("gzip", "-f", fmt.Sprintf("%s/patch.tar", tempDir))

	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func TarGz(files []string, outputFile string) error {
	// Create the output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	// Create a gzip writer
	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	// Create a tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over the input files
	for _, file := range files {
		err = addFileToTar(file, tw)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToTar(file string, tw *tar.Writer) error {
	// Open the input file
	inFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// Get the file information
	info, err := inFile.Stat()
	if err != nil {
		return err
	}

	// Create a tar header based on the file info
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	// Set the name of the file within the tar archive
	header.Name = filepath.Base(file)

	// Write the header the tar writer
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy the file content to the tar writer
	_, err = io.Copy(tw, inFile)
	if err != nil {
		return err
	}

	return nil
}
