#!/bin/sh

crontab /crontab.txt
echo "starting cronjob" > /dev/stdout
crond -f -l 5 > /dev/stdout
