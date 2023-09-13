build:
	go build -o bin/update-toolbox main.go

run:
	build && bin/update-toolbox

generate:
	build && bin/update-toolbox patch create package.json