#!/bin/bash

. .env
psql -Atx $TIMESCALE_DATABASE_URL_INIT -f ./createDb.sql
psql -Atx $TIMESCALE_DATABASE_URL -f ./createTable.sql
