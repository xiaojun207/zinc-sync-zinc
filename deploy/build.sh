#!/bin/bash
# author xiaojun207

DOCKER_BASE_REPO="xiaojun207"
APP_NAME="zinc-sync-zinc"

VERSION=$(git describe --tags `git rev-list --tags --max-count=1`)
VERSION=${VERSION/v/} # eg.: 0.2.5
DOCKER_BUILD_TAG="${VERSION}"

docker build \
  --build-arg VERSION="${VERSION}" \
   -t ${DOCKER_BASE_REPO}/${APP_NAME}:${VERSION} \
   -f deploy/Dockerfile .

#docker tag ${DOCKER_BASE_REPO}/${APP_NAME}:${VERSION} ${DOCKER_BASE_REPO}/${APP_NAME}:latest
#docker push ${DOCKER_BASE_REPO}/${APP_NAME}:${VERSION}
#docker push ${DOCKER_BASE_REPO}/${APP_NAME}:latest
