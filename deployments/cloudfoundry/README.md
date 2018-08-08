# golang-url-shortener on cloudfoundry

## configuration 

1. Either compile or download the linux amd64 binary and copy it into this directory (e.g. `cp ../../releases/golang-url-shortener_linux_amd64/golang-url-shortener .` if you compiled it yourself via make)
1. `cp manifest-example.yml manifest.yml` and edit to meet your needs
1. (optional) create any services that may be required for securing env variables or things like redis, for example:
  * creating a cups service to hold oauth keys: `cf create-user-provided-service gourl-oauth -p '{"githubClientID":"<some id>","githubClientSecret":"<some key>"}'`
  * creating a redis service for later binding: `cf create-service p-redis default gourl-redis-service`
1. (optional) modify run.sh to set `REDIS_SERVICE_NAME` to match the name of the redis service for your cloudfoundry implementation

## deployment

`cf push` or `cf push <custom app name>`