#!/bin/bash

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

echo "[stop] stopping 4 oracles"

kill $(cat $LOG_DIR/oracle0.pid) &>/dev/null
rm $LOG_DIR/oracle0.pid &>/dev/null

kill $(cat $LOG_DIR/oracle1.pid) &>/dev/null
rm $LOG_DIR/oracle1.pid &>/dev/null

kill $(cat $LOG_DIR/oracle2.pid) &>/dev/null
rm $LOG_DIR/oracle2.pid &>/dev/null

kill $(cat $LOG_DIR/oracle3.pid) &>/dev/null
rm $LOG_DIR/oracle3.pid &>/dev/null

echo "[done]"
