#!/bin/bash

. .env
psql -Atx $TIMESCALE_DATABASE_URL -f ./createDb.sql
