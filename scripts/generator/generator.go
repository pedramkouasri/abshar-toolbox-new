package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
	"github.com/pedramkousari/abshar-toolbox-new/types"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type patchService struct {
	cnf config.Config
}

func NewPatchService(cnf config.Config) patchService {
	return patchService{
		cnf: cnf,
	}
}

func (us patchService) Handle(packagePathFile string) error {
	if !utils.FileExists(packagePathFile) {
		return fmt.Errorf("File Not Exists is Path: %s", packagePathFile)
	}

	file, err := os.Open(packagePathFile)
	if err != nil {
		return fmt.Errorf("Can Not open file in path %s - %v", file, err)
	}

	pkg := []types.Packages{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pkg)
	if err != nil {
		return fmt.Errorf("Can not Decode package.json %v", err)
	}

	version := pkg[len(pkg)-1].Version

	diffPackages := utils.GetPackageDiff(pkg)
	if len(diffPackages) == 0 {
		return fmt.Errorf("Not Found Diff Packages")
	}

	var services []string
	for _, dp := range diffPackages {
		services = append(services, dp.ServiceName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), us.cnf.UpdateTimeOut)
	defer cancel()

	wg := new(sync.WaitGroup)
	loading := loading.NewLoading(services, wg)
	hasError := make(chan error)

	for _, pac := range diffPackages {
		if pac.ServiceName == "baadbaan" {
			bs := baadbaan.NewGenerator(us.cnf, pac.Tag1, pac.Tag2, loading)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := bs.Generate(ctx); err != nil {
					hasError <- err
				}
			}()
		}
	}

	go func() {
		wg.Wait()
		close(hasError)
	}()

	for {
		select {
		case res, ok := <-hasError:
			if !ok {
				if err := exportPatch(version, packagePathFile, us.cnf); err != nil {
					return err
				}

				fmt.Println("\nCompleted :)")
				return nil
			}

			if res != nil {
				return fmt.Errorf("Recived Error: %v", res)
			}

		case <-ctx.Done():
			return fmt.Errorf("Time Out Update With Error %v", ctx.Err().Error())
		}
	}

}

func exportPatch(version string, packagePathFile string, cnf config.Config) error {
	tempBuildPath := "./temp/builds"
	entries, err := os.ReadDir(tempBuildPath)
	if err != nil {
		return err
	}

	files := []string{packagePathFile}
	for _, e := range entries {
		files = append(files, tempBuildPath+"/"+e.Name())
	}

	err = os.Mkdir("./builds", 0755)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}

	outputFile := fmt.Sprintf("./builds/%s.tar.gz", version)

	if err := utils.TarGz(files, outputFile); err != nil {
		return err
	}

	if err := utils.EncryptFile([]byte(cnf.EncryptKey), outputFile, outputFile+".enc"); err != nil {
		return err
	}

	if err := os.Remove(outputFile); err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	return nil
}
