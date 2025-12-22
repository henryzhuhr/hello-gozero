#!/bin/bash
# 预先构建项目镜像的脚本，加快 docker compose up 的速度

IMAGE_TAG=1.0.0

GO_VERSION=1.25.5
UV_VERSION=0.9.18
MIRRORS_URL="mirrors.ustc.edu.cn"
CLEAN_APT_CACHE=1


# 镜像列表（格式：镜像名:标签）
IMAGES=(
  "ubuntu:24.04"
  "golang:${GO_VERSION}"
  "ghcr.io/astral-sh/uv:${UV_VERSION}"
  "mysql:9.5"
  "redis:8.4"
  "apache/kafka:4.1.1"
)

for IMAGE in "${IMAGES[@]}"; do
  NAME=$(echo "${IMAGE}" | cut -d: -f1)
  TAG=$(echo "${IMAGE}" | cut -d: -f2-)
  if ! docker images | grep -q "^${NAME}[[:space:]]\+${TAG}[[:space:]]"; then
    echo "pull image: ${IMAGE}"
    docker pull "${IMAGE}" || {
      echo "failed to pull image ${IMAGE}, aborting!";
      exit 1;
    }
  else
    echo "found ${IMAGE}, skip docker pull."
  fi
done

docker build -t hello-gozero:${IMAGE_TAG} -f dockerfiles/Dockerfile \
  --build-arg GO_VERSION=${GO_VERSION} \
  --build-arg UV_VERSION=${UV_VERSION} \
  --build-arg MIRRORS_URL=${MIRRORS_URL} \
  --build-arg CLEAN_APT_CACHE=${CLEAN_APT_CACHE} \
  --no-cache .