all: buildNodeFrontend getCMDDependencies embedFrontend getGoDependencies runUnitTests buildProject

runUnitTests:
	go test -v ./...

buildNodeFrontend:
	cd static && yarn install
	cd static && yarn build
	cd static && rm build/static/**/*.map

embedFrontend:
	cd handlers/tmpls && esc -o tmpls.go -pkg tmpls -include ^*\.html .
	cd handlers && esc -o static.go -pkg handlers -prefix ../static/build ../static/build
	bash build/info.sh

getCMDDependencies:
	go get -v github.com/mattn/goveralls
	go get -v github.com/mjibson/esc
	go get -v github.com/mitchellh/gox

getGoDependencies:
	go get -v ./...

buildProject:
	rm -rf releases 
	mkdir releases
	gox -output="releases/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -osarch="linux/amd64 linux/arm windows/amd64 windows/386"
	find releases -maxdepth 1 -mindepth 1 -type d -exec cp build/config.yaml {} \;
	find releases -maxdepth 1 -mindepth 1 -type d -exec tar -cvjf {}.tar.bz2 {} \;

buildDockerImage:
	rm -rf docker_releases 
	mkdir docker_releases
	CGO_ENABLED=0 gox -output="docker_releases/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -osarch="linux/amd64 linux/arm"
	docker build -t maxibanki/golang_url_shortener:arm -f Dockerfile.arm .
	docker build -t maxibanki/golang_url_shortener -f Dockerfile.amd64 .