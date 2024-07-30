#!/bin/bash

. .env
pnpm i
pnpm prisma generate
pnpm prisma db push
cp .env $DEST
cp docker/production/docker-compose.yml $DEST
cp docker/production/Dockerfile $DEST
cp build/metrics-backend $DEST
rm -rf $DEST/views
cp -r views $DEST
cd $DEST
docker-compose up -d --build
