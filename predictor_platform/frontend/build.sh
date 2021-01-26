#/bin/sh
set -xe
npm run build:prod
cp -r  dist ../server/frontend/
