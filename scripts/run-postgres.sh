#!/usr/bin/env bash

docker run --rm \
    -d \
    -p 5432:5432 \
    --name tagsdb \
    -e POSTGRES_PASSWORD=supersecret \
    postgres:16.4
