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
	src string
}

func NewUpdateService(cnf config.Config) updateService {
	return updateService{
		cnf: cnf,
	}
}

func (us updateService) Handle(ctx context.Context, resChan chan bool) {
	wg := new(sync.WaitGroup)
	loading := loading.NewLoading([]string{"baadbaan", "XXX"}, wg)
	defer wg.Wait()

	us.cnf.SetStartTime()
	bs := baadbaan.NewBaadbaan(us.cnf, "15-10")

	chanComplete := make(chan struct{})
	go func() {
		bs.Update(loading)
		chanComplete <- struct{}{}
	}()

	select {
	case <-chanComplete:
		resChan <- true
		fmt.Println("completed")

	case <-ctx.Done():
		resChan <- false
	}

}
