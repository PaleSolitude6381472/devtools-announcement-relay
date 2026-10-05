#!/bin/sh
set -eu

curl --fail-with-body \
  --request POST \
  --header 'Content-Type: application/json' \
  --data '{"kind":"build_completed","summary":"linux-amd64 build passed","details":"revision a1b2c3"}' \
  http://localhost:8080/announcements
