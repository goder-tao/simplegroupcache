#!/bin/bash
# trap is similar with go defer
trap "rm server; kill 0" EXIT

go build -o server
./server -port=8001 -api=false &
./server -port=8002 -api=false &
./server -port=8003 -api=true  &

sleep 2
echo ">>> test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &

wait
