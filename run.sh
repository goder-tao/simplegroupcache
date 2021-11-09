#!/bin/bash
# trap is similar with go defer
trap "rm server; kill 0" EXIT

go build -o server
./server -ip=localhost -port=8101 -api=true -register=true  &
./server -ip=localhost -port=8102 &
./server -ip=localhost -port=8103 &

sleep 2
echo ">>> test"
curl "http://localhost:12000/api?key=Tom" &
sleep 1 &
curl "http://localhost:12000/api?key=Tom" &
sleep 1 &
curl "http://localhost:12000/api?key=Tom" &
sleep 1 &
curl "http://localhost:12000/api?key=Jack" &
sleep 1 &
curl "http://localhost:12000/api?key=Jack" &
sleep 1 &
curl "http://localhost:12000/api?key=Sam" &

wait
