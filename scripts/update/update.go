package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/types"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type updateService struct {
	cnf config.Config
}

func NewUpdateService(cnf config.Config) updateService {
	return updateService{
		cnf: cnf,
	}
}

func (us updateService) Handle(fileSrc string) error {
	if err := os.Mkdir("./temp", 0755); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("create directory err: %s", err)
		}
	}
	logger.Info("Created Temp Directory")

	if err := utils.DecryptFile([]byte(us.cnf.EncryptKey), fileSrc, strings.TrimSuffix(fileSrc, ".enc")); err != nil {
		return fmt.Errorf("Decrypt File err: %s", err)
	}

	logger.Info("Decrypted File")

	if err := utils.UntarGzip(strings.TrimSuffix(fileSrc, ".enc"), "./temp"); err != nil {
		return fmt.Errorf("UnZip File err: %s", err)
	}
	logger.Info("UnZiped File")

	packagePathFile := "./temp/package.json"

	if _, err := os.Stat(packagePathFile); err != nil {
		return fmt.Errorf("package.json is err: %s", err)
	}

	logger.Info("Exists package.json")

	file, err := os.Open(packagePathFile)
	if err != nil {
		return fmt.Errorf("open package.json is err: %s", err)
	}
	logger.Info("Opened package.json")

	pkg := []types.Packages{}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pkg)
	if err != nil {
		return fmt.Errorf("decode package.json is err: %s", err)
	}

	logger.Info("Decode package.json")

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

	us.cnf.SetStartTime()
	wg := new(sync.WaitGroup)
	loading := loading.NewLoading(services, wg)
	hasError := make(chan error)

	for _, pac := range diffPackages {
		if pac.ServiceName == "baadbaan" {
			bs := baadbaan.NewUpdate(us.cnf, "15-10", loading)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := bs.Update(ctx); err != nil {
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
