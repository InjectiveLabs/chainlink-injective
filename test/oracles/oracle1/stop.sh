#!/bin/bash

CWD=$(pwd)
LOG_DIR="${LOG_DIR:-$CWD/var/oracles}"

echo "[stop] stopping oracle1"

kill $(cat $LOG_DIR/oracle1.pid) &>/dev/null
rm $LOG_DIR/oracle1.pid &>/dev/null

echo "[done]"
