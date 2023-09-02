package update

import (
	"context"
	"fmt"
	"time"
)

type updateService struct{}

func NewUpdateService() updateService {
	return updateService{}
}

func (us updateService) Handle(ctx context.Context, resChan chan bool) {
	fmt.Println("Start")

	select {
	case <-time.After(time.Second * 20):
		resChan <- true
		fmt.Println("Stopped")

	case <-ctx.Done():
		resChan <- false
	}

}
