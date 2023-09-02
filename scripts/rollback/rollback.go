package rollback

import (
	"context"
	"fmt"
	"time"
)

type rollbackService struct{}

func NewUpdateService() rollbackService {
	return rollbackService{}
}

func (us rollbackService) Handle(ctx context.Context) {
	fmt.Println("Started")
	<-time.After(time.Second * 30)
	fmt.Println("Stopped")
}
