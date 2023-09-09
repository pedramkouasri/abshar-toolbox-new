package rollback

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/db"
)

type rollbackService struct {
	cnf config.Config
}

func NewRollbackService(cnf config.Config) rollbackService {
	return rollbackService{
		cnf: cnf,
	}
}

func (rb rollbackService) Handle() error {
	wg := new(sync.WaitGroup)

	hasError := make(chan bool)

	ctx, cancel := context.WithTimeout(context.Background(), rb.cnf.RollbackTimeOut)
	defer cancel()

	p, _ := strconv.Atoi(string(db.NewBoltDB().Get("baadbaan")))
	bs := baadbaan.NewBaadbaanRollback(rb.cnf, "15-10", p)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := bs.Rollback(ctx); err != nil {
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
				return fmt.Errorf("Recived Error In Rollback")
			}

		case <-ctx.Done():
			return fmt.Errorf("Time Out Restore With Error %v", ctx.Err().Error())
		}
	}

}
