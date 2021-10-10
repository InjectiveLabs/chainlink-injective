#!/bin/bash

source ./test/scripts/common.sh

set -e

echo "Starting job '$1' ($2) via Chainlink node"

CL_URL="http://localhost:6688"

login_cl "$CL_URL"

payload=$(
  jq --arg f "$(cat $2)" '.toml = $f' < ./test/scripts/jobspec_template.json
)

OUTPUT=$(curl -s -b ./.cookiefile -d "$payload" -X POST -H 'Content-Type: application/toml' "$CL_URL/v2/jobs")

echo "Job has been added via Chainlink node"
echo "Done adding Job '$1'"
