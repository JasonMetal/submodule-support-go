#!/bin/bash
cd /data/docker/services-release-test
myversion=$2
myproject=$1
gitlab=$3
rm -rf $myproject
git clone git@gitee.com:DXTeam/$myproject.git
cd $myproject
#git checkout release
#git submodule update --init --recursive
#git submodule init
#git submodule sync --recursive
#git submodule update --remote
test="export GO111MODULE=on && cd /go/src/$myproject && CGO_ENABLED=1 GOOS=linux GOARCH=amd64; go mod tidy; go build  -o 'http-server' main.go;"
docker run --rm -v /data/docker/services-release-test/$myproject:/go/src/$myproject -v /data1/docker-go-build/mod:/go/pkg/mod  alpine:latest
latest  bash -c "$test"
