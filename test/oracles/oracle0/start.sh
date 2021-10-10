#!/bin/bash

set -e

# enables ENV export when sourcing
set -o allexport

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

if [[ -f $LOG_DIR/oracle0.pid ]]
then
  echo "[cleanup] oracle0 started already, stopping"
  kill $(cat $LOG_DIR/oracle0.pid) &>/dev/null
  rm $LOG_DIR/oracle0.pid
fi

echo "[start] running oracle0"
mkdir -p $LOG_DIR

source ./test/oracles/oracle0/env
source ./test/oracles/oracle0/external_initiator_injective-ei0.env
injective-ocr2 start > $LOG_DIR/oracle0.log 2>&1 &
echo $! >$LOG_DIR/oracle0.pid

echo "[post-start]"
echo "Logs:"
echo "  * tail -f ./var/oracles/oracle0.log"
echo
echo "Stopping:"
echo "  * ./test/oracles/oracle0/stop.sh"
