#!/bin/bash
# trap is similar with go defer
trap "rm server; kill 0" EXIT

go build -o server
./server -ip=localhost -port=8101 -api=true -register=true  &
sleep 3 
./server -ip=localhost -port=8102 &
./server -ip=localhost -port=8103 &

sleep 3
echo ">>> test"
curl "http://localhost:12000/api?key=Tom" 
curl "http://localhost:12000/api?key=Tom" 
curl "http://localhost:12000/api?key=Jack" 
curl "http://localhost:12000/api?key=Jack" 
curl "http://localhost:12000/api?key=Sam" 
curl "http://localhost:12000/api?key=Sam" 

wait
