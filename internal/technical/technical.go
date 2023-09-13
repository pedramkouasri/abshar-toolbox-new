package technical

import (
	"context"
	"fmt"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type technical struct {
	tempDir       string
	outFile       string
	dir           string
	branch        string
	tag1          string
	tag2          string
	serviceName   string
	containerName string
	env           *utils.ConfigService
	percent       int
	loading       contracts.Loader
	cnf           config.Config
}

func (t *technical) exec(ctx context.Context, percent int, message string, fn func() error) (err error) {
	if err = ctx.Err(); err != nil {
		return fmt.Errorf("context is err :%v", err)
	}

	if err = fn(); err != nil {
		return err
	}

	t.setPercent(percent)
	logger.Info(message)
	return
}

func (t *technical) setPercent(percent int) {
	t.percent = percent
	t.loading.Update(t.serviceName, t.percent)
}
