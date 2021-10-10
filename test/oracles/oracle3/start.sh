#!/bin/bash

set -e

# enables ENV export when sourcing
set -o allexport

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

if [[ -f $LOG_DIR/oracle3.pid ]]
then
  echo "[cleanup] oracle3 started already, stopping"
  kill $(cat $LOG_DIR/oracle3.pid) &>/dev/null
  rm $LOG_DIR/oracle3.pid
fi

echo "[start] running oracle3"
mkdir -p $LOG_DIR

source ./test/oracles/oracle3/env
source ./test/oracles/oracle3/external_initiator_injective-ei3.env
injective-ocr2 start > $LOG_DIR/oracle3.log 2>&1 &
echo $! >$LOG_DIR/oracle3.pid

echo "[post-start]"
echo "Logs:"
echo "  * tail -f ./var/oracles/oracle3.log"
echo
echo "Stopping:"
echo "  * ./test/oracles/oracle3/stop.sh"
