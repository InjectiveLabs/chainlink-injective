#!/bin/bash

set -e

# enables ENV export when sourcing
set -o allexport

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

if [[ -f $LOG_DIR/oracle1.pid ]]
then
  echo "[cleanup] oracle1 started already, stopping"
  kill $(cat $LOG_DIR/oracle1.pid) &>/dev/null
  rm $LOG_DIR/oracle1.pid
fi

echo "[start] running oracle1"
mkdir -p $LOG_DIR

source ./test/oracles/oracle1/env
source ./test/oracles/oracle1/external_initiator_injective-ei1.env
injective-ocr2 start > $LOG_DIR/oracle1.log 2>&1 &
echo $! >$LOG_DIR/oracle1.pid

echo "[post-start]"
echo "Logs:"
echo "  * tail -f ./var/oracles/oracle1.log"
echo
echo "Stopping:"
echo "  * ./test/oracles/oracle1/stop.sh"
