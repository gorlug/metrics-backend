#!/bin/bash

. .env
psql -Atx $DATABASE_URL -c "insert into users (email) values ('$1');"
