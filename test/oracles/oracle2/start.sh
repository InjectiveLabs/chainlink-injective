#!/bin/bash

set -e

# enables ENV export when sourcing
set -o allexport

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

if [[ -f $LOG_DIR/oracle2.pid ]]
then
  echo "[cleanup] oracle2 started already, stopping"
  kill $(cat $LOG_DIR/oracle2.pid) &>/dev/null
  rm $LOG_DIR/oracle2.pid
fi

echo "[start] running oracle2"
mkdir -p $LOG_DIR

source ./test/oracles/oracle2/env
source ./test/oracles/oracle2/external_initiator_injective-ei2.env
injective-ocr2 start > $LOG_DIR/oracle2.log 2>&1 &
echo $! >$LOG_DIR/oracle2.pid

echo "[post-start]"
echo "Logs:"
echo "  * tail -f ./var/oracles/oracle2.log"
echo
echo "Stopping:"
echo "  * ./test/oracles/oracle2/stop.sh"
