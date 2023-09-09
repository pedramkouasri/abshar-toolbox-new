package update

import (
	"context"
	"fmt"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
)

type updateService struct {
	cnf config.Config
}

func NewUpdateService(cnf config.Config) updateService {
	return updateService{
		cnf: cnf,
	}
}

func (us updateService) Handle() error {
	wg := new(sync.WaitGroup)

	loading := loading.NewLoading([]string{"baadbaan"}, wg)

	us.cnf.SetStartTime()
	bs := baadbaan.NewBaadbaan(us.cnf, "15-10", loading)

	hasError := make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), us.cnf.UpdateTimeOut)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bs.Update(ctx); err != nil {
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
