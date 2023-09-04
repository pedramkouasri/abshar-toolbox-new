package update

import (
	"context"
	"fmt"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
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
	us.cnf.SetStartTime()
	bs := baadbaan.NewBaadbaan(us.cnf, "15-10")

	chanComplete := make(chan struct{})
	go func() {
		bs.Update()
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
