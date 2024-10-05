#!/bin/sh

crontab /crontab.txt
crond -f -l 5
