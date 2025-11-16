#!/bin/bash
set -e

scp ./docker-compose.yml user@192.168.1.1:/home/user/smart-control
scp ./Dockerfile user@192.168.1.1:/home/user/smart-control
ssh user@192.168.1.1 'cd ~/smart-control && docker compose -f docker-compose.yml build --pull smart-control && docker compose -f docker-compose.yml up -d --force-recreate smart-control && docker logs smart-control'
