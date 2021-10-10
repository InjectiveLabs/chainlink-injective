#!/bin/bash

source ./test/scripts/common.sh

set -e

echo "Adding Bridge '$1' ($2) to Chainlink node"

CL_URL="http://localhost:6688"

login_cl "$CL_URL"

payload=$(
  cat <<EOF
{
"name": "$1",
"url": "$2/runs"
}
EOF
)

curl -s -b ./.cookiefile -d "$payload" -X POST -H 'Content-Type: application/json' "$CL_URL/v2/bridge_types" &>/dev/null

echo "Bridge has been added to Chainlink node"
echo "Done adding Bridge '$1'"
