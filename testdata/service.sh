#!/bin/bash
pid=
echo "service.sh running as child of PID $$"
trap 'echo "trapped SIGTERM for $pid"; exit 15' SIGTERM
trap 'echo "Pid $pid exited"' EXIT
sleep 10000 & pid=$!
wait
pid=
