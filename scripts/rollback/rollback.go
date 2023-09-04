package rollback

import (
	"context"
	"time"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
)

type rollbackService struct {
	cnf config.Config
}

func NewRollbackService(cnf config.Config) rollbackService {
	return rollbackService{
		cnf: cnf,
	}
}

func (us rollbackService) Handle(ctx context.Context, resChan chan bool) (err error) {
	defer func() {
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	logger.Info("Rollback Started")
	<-time.After(time.Second * 5)
	logger.Info("Rollback Finished")

	return
}
