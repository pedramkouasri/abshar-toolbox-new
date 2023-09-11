package update

import (
	"context"
	"fmt"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
	"github.com/pedramkousari/abshar-toolbox-new/types"
)

type updateService struct {
	cnf config.Config
}

func NewUpdateService(cnf config.Config) updateService {
	return updateService{
		cnf: cnf,
	}
}

func (us updateService) Handle(diffPackages []types.CreatePackageParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), us.cnf.UpdateTimeOut)
	defer cancel()

	us.cnf.SetStartTime()
	wg := new(sync.WaitGroup)
	hasError := make(chan error)

	var services []string
	for _, dp := range diffPackages {
		services = append(services, dp.ServiceName)
	}
	loading := loading.NewLoading(services, wg)

	for _, pac := range diffPackages {
		if pac.ServiceName == "baadbaan" {
			bs := baadbaan.NewUpdate(us.cnf, pac.Tag2, loading)

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
