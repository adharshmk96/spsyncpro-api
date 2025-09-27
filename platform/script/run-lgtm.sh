#!/bin/bash

RELEASE=${1:-latest}

docker pull docker.io/grafana/otel-lgtm:"${RELEASE}"

docker rm lgtm -f

docker run \
	--name lgtm \
    -d \
	-p 3000:3000 \
	-p 4317:4317 \
	-p 4318:4318 \
	--rm \
	-ti \
	-v "$PWD"/_container/grafana:/data/grafana \
	-v "$PWD"/_container/prometheus:/data/prometheus \
	-v "$PWD"/_container/loki:/data/loki \
	-e GF_PATHS_DATA=/data/grafana \
	--env-file .env \
	docker.io/grafana/otel-lgtm:"${RELEASE}"