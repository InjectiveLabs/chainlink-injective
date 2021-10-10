#!/bin/bash

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

echo "[stop] stopping oracle0"

kill $(cat $LOG_DIR/oracle0.pid) &>/dev/null
rm $LOG_DIR/oracle0.pid &>/dev/null

echo "[done]"
