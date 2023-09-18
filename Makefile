build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/update-toolbox main.go

build-server: build
	scp bin/update-toolbox root@10.10.10.217:/var/www/abshar/bin

build-229: build
	scp bin/update-toolbox root@10.10.10.229:/var/www/html/baadbaan-docker/services/update-toolbox	

run: build
	bin/update-toolbox

generate: build
	bin/update-toolbox patch create package.json