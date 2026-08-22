#!/bin/sh
set -eu
docker buildx build --platform linux/amd64 -t benzhi/groundops:test .
