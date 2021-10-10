#!/bin/bash

set -e

sleep 1
./test/scripts/cl_add_ei.sh injective-ei0 http://host.docker.internal:8866
sleep 1
./test/scripts/cl_add_ei.sh injective-ei1 http://host.docker.internal:8867
sleep 1
./test/scripts/cl_add_ei.sh injective-ei2 http://host.docker.internal:8868
sleep 1
./test/scripts/cl_add_ei.sh injective-ei3 http://host.docker.internal:8869

mv ./external_initiator_injective-ei0.env ./test/oracles/oracle0
mv ./external_initiator_injective-ei1.env ./test/oracles/oracle1
mv ./external_initiator_injective-ei2.env ./test/oracles/oracle2
mv ./external_initiator_injective-ei3.env ./test/oracles/oracle3
