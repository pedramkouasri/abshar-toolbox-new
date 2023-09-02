build:
	go build -o bin/abshar cmd/server/main.go

run:
	build && bin/abshar