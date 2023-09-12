package toolbox

import (
	"context"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type toolbox struct {
	tempDir     string
	outFile     string
	dir         string
	branch      string
	tag1        string
	tag2        string
	serviceName string
	env         *utils.ConfigService
	percent     int
	loading     contracts.Loader
	cnf         config.Config
}

func (t *toolbox) exec(ctx context.Context, percent int, message string, fn func() error) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}

	if err = fn(); err != nil {
		return
	}

	t.setPercent(percent)
	logger.Info(message)
	return
}

func (t *toolbox) setPercent(percent int) {
	t.percent = percent
	t.loading.Update(t.serviceName, t.percent)
}
