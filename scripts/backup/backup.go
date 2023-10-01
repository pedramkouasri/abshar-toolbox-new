package backup

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type backupService struct {
	cnf config.Config
}

func NewBackupService(cnf config.Config) backupService {
	return backupService{
		cnf: cnf,
	}
}

func (us backupService) Handle(branchName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), us.cnf.UpdateTimeOut)
	defer cancel()

	wg := new(sync.WaitGroup)

	services := []string{"baadbaan"}
	if utils.DirectoryExists(us.cnf.DockerComposeDir + "/services/technical-risk-micro-service") {
		services = append(services, "technical")
	}

	if utils.DirectoryExists(us.cnf.DockerComposeDir + "/services/asset-discovery") {
		services = append(services, "discovery")
	}

	loading := loading.NewLoading(services, wg)
	hasError := make(chan error)

	for _, serviceName := range services {
		if serviceName == "baadbaan" {
			bs := baadbaan.NewBackup(us.cnf, branchName, loading)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := bs.Backup(ctx); err != nil {
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
				if err := exportPatch(branchName, us.cnf); err != nil {
					return err
				}

				fmt.Println("\nCompleted :)")
				return nil
			}

			if res != nil {
				return fmt.Errorf("Recived Error: %v", res)
			}

		case <-ctx.Done():
			return fmt.Errorf("Time Out Backup With Error %v", ctx.Err().Error())
		}
	}

}

func exportPatch(version string, cnf config.Config) error {
	tempBuildPath := "./temp/builds"

	os.WriteFile(tempBuildPath+"/branch.txt", []byte(version), 0644)

	entries, err := os.ReadDir(tempBuildPath)
	if err != nil {
		return err
	}

	files := []string{}
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

	backupDir := fmt.Sprintf("%s/baadbaan_new/storage/app/backup/", cnf.DockerComposeDir)
	if utils.DirectoryExists(backupDir) == false {
		if err := os.Mkdir(backupDir, 755); err != nil {
			return fmt.Errorf("Can not Backup Dir in Baadbaan %v", err)
		}

		if err := utils.ChangePermision("www-data", backupDir); err != nil {
			return fmt.Errorf("Can not Change permission Backup Dir in Baadbaan %v", err)
		}
	}

	if err := os.Rename(outputFile, fmt.Sprintf("%s/%s.tar.gz", backupDir, version)); err != nil {
		return fmt.Errorf("Cannot Move File err is: %s", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Println("Error deleting file:", err)
			return err
		}
	}

	if err := os.Remove(tempBuildPath); err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	return nil
}
