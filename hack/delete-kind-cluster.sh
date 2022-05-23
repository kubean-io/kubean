#!/usr/bin/env bash

set -o errexit
set -o nounset

list=$(docker ps|grep kubean|grep -E 'hours|day|hour|days'|awk '{print $1}' && docker ps|grep kubean|grep -E '.*minutes'|awk '$4>=10{print $1}')
for i in $list
do
  docker rm -f $i
done