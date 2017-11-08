all: buildNodeFrontend getCMDDependencies embedFrontend getGoDependencies runUnitTests buildProject

runUnitTests:
	go test -v ./...
	goveralls -service=travis-ci

buildNodeFrontend:
	@cd static && yarn install
	@cd static && yarn build
	@cd static && rm build/static/**/*.map

embedFrontend:
	@cd handlers/tmpls && esc -o tmpls.go -pkg tmpls -include ^*\.tmpl .
	@cd handlers && esc -o static.go -pkg handlers -prefix ../static/build ../static/build

getCMDDependencies:
	go get -v github.com/mattn/goveralls
	go get -v github.com/mjibson/esc
	go get -v github.com/mitchellh/gox

getGoDependencies:
	go get -v ./...

buildProject:
	@cd build && go build schema.go && ./schema
	@mkdir releases
	gox -output="releases/{{.OS}}_{{.Arch}}/{{.OS}}_{{.Arch}}"
	find releases -maxdepth 1 -mindepth 1 -type d -exec cp build/config.json {} \;
	find releases -maxdepth 1 -mindepth 1 -type d -exec cp build/schema.json {} \;
	find releases -maxdepth 1 -mindepth 1 -type d -exec tar -cvjf {}.tar.bz2 {} \;
