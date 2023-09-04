package update

import (
	"context"
	"fmt"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
)

type updateService struct {
	cnf config.Config
}

func NewUpdateService(cnf config.Config) updateService {
	return updateService{
		cnf: cnf,
	}
}

func (us updateService) Handle(ctx context.Context, resChan chan bool) {
	wgL := new(sync.WaitGroup)
	loading := loading.NewLoading([]string{"baadbaan", "XXX"}, wgL)
	defer wgL.Wait()

	us.cnf.SetStartTime()
	bs := baadbaan.NewBaadbaan(us.cnf, "15-10", loading)

	perServiceChan := make(chan bool)
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		bs.Update(ctx, perServiceChan, false)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		bs.Update(ctx, perServiceChan, true)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		bs.Update(ctx, perServiceChan, true)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		bs.Update(ctx, perServiceChan, true)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		bs.Update(ctx, perServiceChan, true)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		bs.Update(ctx, perServiceChan, true)
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(perServiceChan)
	}()

	for {
		select {
		case res, ok := <-perServiceChan:
			fmt.Println(res)

			if !ok {
				logger.Info("Complete update all service")
				resChan <- true
				return
			} else if res == false {
				resChan <- false
				return
			}

		case <-ctx.Done():
			resChan <- false
			return
		}
	}

}
