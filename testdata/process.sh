#!/bin/bash
echo "process.sh running as child of PID $$"
for ((i=3; i>=1; i--))
do
    sleep .5
done
