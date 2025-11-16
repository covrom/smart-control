#!/bin/bash
set -e

(cd ./win && GOOS=windows go build -o ../dist/SMARTDataCollector.exe . )
cp ./win/script.nsi ./dist/
docker compose -f docker-compose.nsis.yml up
