#!/bin/sh

REDIS_SERVICE_NAME="thd-redis"

CUPS=$(echo $VCAP_SERVICES | grep "user-provided")
REDIS=$(echo $VCAP_SERVICES | grep "$REDIS_SERVICE_NAME")

if [ "$CUPS" != "" ]; then
    export GUS_GITHUB_CLIENT_ID="$(echo $VCAP_SERVICES | jq -r '.["'user-provided'"][0].credentials.githubClientID')"
    export GUS_GITHUB_CLIENT_SECRET="$(echo $VCAP_SERVICES | jq -r '.["'user-provided'"][0].credentials.githubClientSecret')"
fi

if [ "$REDIS" != "" ]; then
    export GUS_REDIS_HOST="$(echo $VCAP_SERVICES | jq -r '.["'$REDIS_SERVICE_NAME'"][0].credentials.host'):$(echo $VCAP_SERVICES | jq -r '.["'$REDIS_SERVICE_NAME'"][0].credentials.port')"
    export GUS_REDIS_PASSWORD="$(echo $VCAP_SERVICES | jq -r '.["'$REDIS_SERVICE_NAME'"][0].credentials.password')"
fi

echo "#### Starting golang-url-shortener..."

./golang-url-shortener 
