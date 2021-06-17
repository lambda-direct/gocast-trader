#!/usr/bin/env bash
set -e

dir=$(dirname "$0")
docker build -t dokku/ticker:latest "$dir/.." -f "$dir/Dockerfile"
docker save dokku/ticker:latest | bzip2 | pv | ssh dan@142.93.134.108 "bunzip2 | docker load"
ssh dan@142.93.134.108 "dokku tags:create ticker previous; dokku tags:deploy ticker latest && dokku tags:create ticker latest"
