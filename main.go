package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func leftOne(ctx context.Context) {
	select {
	case <-time.After(time.Second * 8):
		fmt.Println("1")

	case <-ctx.Done():
		fmt.Println("Timeout Left2")
	}
}

func leftTwo(ctx context.Context) {
	select {
	case <-time.After(time.Second * 2):
		fmt.Println("2")

	case <-ctx.Done():
		fmt.Println("Timeout Left1")
	}
}

func branch(ctx context.Context) {
	go leftOne(ctx)
	go leftTwo(ctx)

	select {
	case <-time.After(time.Second * 4):
		fmt.Println("Brnhc")

	case <-ctx.Done():
		fmt.Println("Timeout Brnach")
	}
}
func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)

	sig := make(chan os.Signal)
	go branch(ctx)

	go func() {
		time.Sleep(time.Second * 3)
		cancel()
	}()

	<-sig

}
