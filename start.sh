#!/bin/sh

set -e

# load enviroment variables to shell
# echo "loading enviroment variables"
# source /app/app.env

echo "run db migration"
echo "DB_SOURCE : " $DB_SOURCE
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"