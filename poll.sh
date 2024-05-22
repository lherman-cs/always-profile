#!/bin/bash

while true; do
	curl "http://localhost:6060/debug/pprof/profile?seconds=5" >/dev/null
	sleep 60
done
