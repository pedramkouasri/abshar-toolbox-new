package restore

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/internal/baadbaan"
	"github.com/pedramkousari/abshar-toolbox-new/internal/discovery"
	"github.com/pedramkousari/abshar-toolbox-new/internal/technical"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/loading"
)

type restoreService struct {
	cnf config.Config
}

func NewRestoreService(cnf config.Config) restoreService {
	return restoreService{
		cnf: cnf,
	}
}

func (us restoreService) Handle(branchName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), us.cnf.UpdateTimeOut)
	defer cancel()

	wg := new(sync.WaitGroup)

	entries, err := os.ReadDir("./temp")
	if err != nil {
		return err
	}

	suffix := ".tar.gz"
	services := []string{}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			services = append(services, strings.TrimSuffix(e.Name(), suffix))
		}
	}

	loading := loading.NewLoading(services, wg)
	hasError := make(chan error)

	for _, serviceName := range services {
		if serviceName == "baadbaan" {
			bs := baadbaan.NewRestore(us.cnf, branchName, loading)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := bs.Restore(ctx); err != nil {
					hasError <- err
				}
			}()
		}

		if serviceName == "technical" {
			ts := technical.NewRestore(us.cnf, branchName, loading)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := ts.Restore(ctx); err != nil {
					hasError <- err
				}
			}()
		}

		if serviceName == "discovery" {
			ds := discovery.NewRestore(us.cnf, branchName, loading)

			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := ds.Restore(ctx); err != nil {
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
