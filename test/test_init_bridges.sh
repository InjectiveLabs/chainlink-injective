#!/bin/bash

set -e

sleep 1
./test/scripts/cl_add_bridge.sh injective-ea0 http://host.docker.internal:8866
sleep 1
./test/scripts/cl_add_bridge.sh injective-ea1 http://host.docker.internal:8867
sleep 1
./test/scripts/cl_add_bridge.sh injective-ea2 http://host.docker.internal:8868
sleep 1
./test/scripts/cl_add_bridge.sh injective-ea3 http://host.docker.internal:8869
