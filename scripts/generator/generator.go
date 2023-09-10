package generator

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

	wg := new(sync.WaitGroup)

	loading := loading.NewLoading([]string{"baadbaan"}, wg)

	us.cnf.SetStartTime()
	bs := baadbaan.NewGenerator(us.cnf, "15-10", "sss", loading)

	hasError := make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), us.cnf.UpdateTimeOut)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bs.Generate(ctx); err != nil {
			hasError <- true
		}
	}()

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

			if res {
				return fmt.Errorf("Recived Error")
			}

		case <-ctx.Done():
			return fmt.Errorf("Time Out Update With Error %v", ctx.Err().Error())
		}
	}

}
