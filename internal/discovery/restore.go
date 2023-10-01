package discovery

import (
	"context"
	"fmt"
	"path"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

func NewRestore(cnf config.Config, branch string, loading contracts.Loader) *discovery {
	return &discovery{
		dir:         path.Join(cnf.DockerComposeDir, "discovery_new"),
		branch:      branch,
		serviceName: "discovery",
		percent:     0,
		loading:     loading,
		cnf:         cnf,
	}
}

func (b *discovery) Restore(ctx context.Context) error {

	completeSignal := make(chan error)
	go func() {
		defer close(completeSignal)
		if err := b.runRestore(ctx); err != nil {
			completeSignal <- err
		}
	}()

	select {
	case err, ok := <-completeSignal:
		if !ok {
			logger.Info(fmt.Sprintf("Service Restore %s Completed", b.serviceName))
			return nil
		}

		if err != nil {
			return fmt.Errorf("Service Restore Package %s is failed: %v", b.serviceName, err)
		}

		return nil

	case <-ctx.Done():
		logger.Info(fmt.Sprintf("%s Canceled", b.serviceName))
		return ctx.Err()
	}
}

func (d *discovery) runRestore(ctx context.Context) error {
	var err error

	err = d.exec(ctx, 70, "Discovery Restore Code ", func() error {
		return utils.RestoreCode(d.dir)
	})
	if err != nil {
		return fmt.Errorf("Discovery Restore Code Failed Error Is: %s", err)
	}

	err = d.exec(ctx, 100, "Discovery Restore Branch", func() error {
		return utils.SwitchBranch(d.dir, d.branch)
	})
	if err != nil {
		return fmt.Errorf("Discovery Restore Branch Failed Error is: %s", err)
	}

	return nil
}
