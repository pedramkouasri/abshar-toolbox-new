package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func UntarGzip(sourceFile, destinationDir string) error {
	// Open the source gzip file
	gzipFile, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract each file from the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of tar archive
			break
		}
		if err != nil {
			return err
		}

		// Determine the file path for extraction
		target := filepath.Join(destinationDir, header.Name)

		// Create directories if necessary
		if header.Typeflag == tar.TypeDir {
			err := os.MkdirAll(target, 0755)
			if err != nil {
				return err
			}
			continue
		}

		// Create the file for extraction
		file, err := os.Create(target)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file data from the tar entry to the created file
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

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
	file, err := os.Open(fmt.Sprintf("%s/composer-lock-diff.json", tempDir))
	if err != nil {
		return nil, fmt.Errorf("cannot open composer-lock-diff %v", err)
	}
	defer file.Close()

	type ChangesType struct {
		Changes map[string][]string `json:"changes"`
	}

	changesInstance := ChangesType{}

	if err := json.NewDecoder(file).Decode(&changesInstance); err != nil {
		return map[string][]string{}, fmt.Errorf("Cannot Decode composer-lock-diff :%v", err)
	}

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
