package rollback

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/db"
	"github.com/pedramkousari/abshar-toolbox-new/types"
)

type rollbackService struct {
	cnf config.Config
}

func NewRollbackService(cnf config.Config) rollbackService {
	return rollbackService{
		cnf: cnf,
	}
}

func (rb rollbackService) Handle(diffPackages []types.CreatePackageParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), rb.cnf.RollbackTimeOut)
	defer cancel()

	wg := new(sync.WaitGroup)
	hasError := make(chan error)

	for _, pac := range diffPackages {
		if pac.ServiceName == "baadbaan" {
			p, _ := strconv.Atoi(string(db.NewBoltDB().Get("baadbaan")))
			bs := baadbaan.NewRollback(rb.cnf, pac.Tag2, p)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := bs.Rollback(ctx); err != nil {
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
			return fmt.Errorf("Time Out Restore With Error %v", ctx.Err().Error())
		}
	}

}
