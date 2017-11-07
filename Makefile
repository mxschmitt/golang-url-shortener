all: buildNodeFrontend getCMDDependencies embedFrontend getGoDependencies test build uploadCoveralls

test:
	go test -v ./...

buildNodeFrontend:
	@cd static && yarn install
	@cd static && yarn build

embedFrontend:
	@cd handlers/tmpls && esc -o tmpls.go -pkg tmpls -include ^*\.tmpl .
	@cd handlers && esc -o static.go -pkg handlers -prefix ../static/build ../static/build

getCMDDependencies:
	go get -v github.com/mattn/goveralls
	go get -v github.com/mjibson/esc
	go get -v github.com/mitchellh/gox

getGoDependencies:
	go get -v ./...

build:
	@mkdir releases
	gox -output="releases/{{.Dir}}_{{.OS}}_{{.Arch}}"

uploadCoveralls:
	goveralls -service=travis-ci