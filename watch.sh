#!/bin/bash

shopt -s globstar

FILES=$(ls cli/**/*.go proto/*.js web/src/*.ts web/dist/*.{html,css})
inotifywait -m -e close_write $FILES | \
while true; do
  if make all; then
    cd cli && ./smash &
    pid=$!
    read status
    kill $pid
  else
    read status
  fi
  echo $status
done