package discovery

import (
	"context"

	"github.com/pedramkousari/abshar-toolbox-new/config"
	"github.com/pedramkousari/abshar-toolbox-new/contracts"
	"github.com/pedramkousari/abshar-toolbox-new/pkg/logger"
	"github.com/pedramkousari/abshar-toolbox-new/utils"
)

type discovery struct {
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

func (d *discovery) exec(ctx context.Context, percent int, message string, fn func() error) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}

	if err = fn(); err != nil {
		return
	}

	d.setPercent(percent)
	logger.Info(message)
	return
}

func (d *discovery) setPercent(percent int) {
	d.percent = percent
	d.loading.Update(d.serviceName, d.percent)
}
