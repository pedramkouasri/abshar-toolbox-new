build:
	go build -o bin/abshar-toolbox main.go

run:
	build && bin/abshar-toolbox

generate:
	build && bin/abshar-toolbox patch create package.json