#!/usr/bin/env bash

docker exec -i tagsdb \
    psql -d postgres -U postgres <db/ddl/init.sql
