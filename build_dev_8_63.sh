#!/bin/bash
# 专用于 192.168.8.63 机器上的docker编译与部署
set -e

echo "git pull start..."
git pull

echo "please confirm you have done git pull"

echo "latest commit:"
git log -1
sleep 2

echo "working dir:"
pwd
PWD=`pwd`
DOCKER_COMPOSE_DIR=/work/conf/collection-center
cd $DOCKER_COMPOSE_DIR && docker-compose down
echo "stop last version collection-center docker"

echo "cleaning docker cache..."
set +e
docker images | grep none | awk '{print $3}' | xargs -i docker rmi {}
set -e
echo "building docker start..."
cd -
docker build -t collection-center --no-cache=true .

echo "building docker done..."

echo "starting now..."
cd $DOCKER_COMPOSE_DIR && docker-compose up -d

echo "execute done."
