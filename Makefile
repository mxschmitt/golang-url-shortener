all: buildNodeFrontend getCMDDependencies embedFrontend getGoDependencies runUnitTests buildProject

runUnitTests:
	go test -v ./...

buildNodeFrontend:
	cd web && yarn install --ignore-engines
	cd web && yarn build
	cd web && rm build/static/**/*.map

embedFrontend:
	cd cmd/golang-url-shortener && packr2

getCMDDependencies:
	go get -v github.com/mattn/goveralls
	go get -v github.com/gobuffalo/packr/v2/packr2
	go get -v github.com/mitchellh/gox

getGoDependencies:
	go get -v ./...
	# Workaround for: https://github.com/sirupsen/logrus/issues/824
	GOOS=windows go get -v ./...

buildProject:
	rm -rf releases
	mkdir releases
	gox -output="releases/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -osarch="linux/amd64 linux/arm windows/amd64 windows/386 darwin/amd64" -ldflags="-X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagNodeJS=`node --version` -X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagCommit=`git rev-parse HEAD` -X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagYarn=`yarn --version` -X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagCompilationTime=`TZ=UTC date +%Y-%m-%dT%H:%M:%S+0000`" ./cmd/golang-url-shortener
	find releases -maxdepth 1 -mindepth 1 -type d -exec cp config/example.yaml {}/config.yaml \;
	find releases -maxdepth 1 -mindepth 1 -type d -exec tar -cvjf {}.tar.bz2 {} \;

buildDockerImage:
	rm -rf docker_releases
	mkdir docker_releases
	CGO_ENABLED=0 gox -output="docker_releases/{{.Dir}}_{{.OS}}_{{.Arch}}/{{.Dir}}" -osarch="linux/amd64 linux/arm" -ldflags="-X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagNodeJS=`node --version` -X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagCommit=`git rev-parse HEAD` -X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagYarn=`yarn --version` -X github.com/mxschmitt/golang-url-shortener/internal/util.ldFlagCompilationTime=`TZ=UTC date +%Y-%m-%dT%H:%M:%S+0000`" ./cmd/golang-url-shortener
	docker build -t mxschmitt/golang_url_shortener:arm -f build/Dockerfile.arm .
	docker build -t mxschmitt/golang_url_shortener -f build/Dockerfile.amd64 .
