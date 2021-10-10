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

if [[ -f $LOG_DIR/oracle1.pid ]]
then
  echo "[cleanup] oracle1 started already, stopping"
  kill $(cat $LOG_DIR/oracle1.pid) &>/dev/null
  rm $LOG_DIR/oracle1.pid
fi

if [[ -f $LOG_DIR/oracle2.pid ]]
then
  echo "[cleanup] oracle2 started already, stopping"
  kill $(cat $LOG_DIR/oracle2.pid) &>/dev/null
  rm $LOG_DIR/oracle2.pid
fi

if [[ -f $LOG_DIR/oracle3.pid ]]
then
  echo "[cleanup] oracle3 started already, stopping"
  kill $(cat $LOG_DIR/oracle3.pid) &>/dev/null
  rm $LOG_DIR/oracle3.pid
fi

echo "[start] running 4 oracles"
mkdir -p $LOG_DIR

source ./test/oracles/oracle0/env
source ./test/oracles/oracle0/external_initiator_injective-ei0.env
injective-ocr2 start > $LOG_DIR/oracle0.log 2>&1 &
echo $! >$LOG_DIR/oracle0.pid

source ./test/oracles/oracle1/env
source ./test/oracles/oracle1/external_initiator_injective-ei1.env
injective-ocr2 start > $LOG_DIR/oracle1.log 2>&1 &
echo $! >$LOG_DIR/oracle1.pid

source ./test/oracles/oracle2/env
source ./test/oracles/oracle2/external_initiator_injective-ei2.env
injective-ocr2 start > $LOG_DIR/oracle2.log 2>&1 &
echo $! >$LOG_DIR/oracle2.pid

source ./test/oracles/oracle3/env
source ./test/oracles/oracle3/external_initiator_injective-ei3.env
injective-ocr2 start > $LOG_DIR/oracle3.log 2>&1 &
echo $! >$LOG_DIR/oracle3.pid

echo "[post-start]"
echo "Logs:"
echo "  * tail -f ./var/oracles/oracle0.log"
echo "  * tail -f ./var/oracles/oracle1.log"
echo "  * tail -f ./var/oracles/oracle2.log"
echo "  * tail -f ./var/oracles/oracle3.log"
echo
echo "Stopping:"
echo "  * ./test/test_stop_oracles.sh"
