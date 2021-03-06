#!/bin/bash

source ./test/scripts/common.sh

set -e

echo "Adding External Initiator '$1' ($2) to Chainlink node..."

CL_URL="http://localhost:6688"

login_cl "$CL_URL"

payload=$(
  cat <<EOF
{
"name": "$1",
"url": "$2/jobs"
}
EOF
)

result=$(curl -s -b ./.cookiefile -d "$payload" -X POST -H 'Content-Type: application/json' "$CL_URL/v2/external_initiators")

EI_IC_ACCESSKEY=$(jq -r '.data.attributes.incomingAccessKey' <<<"$result")
EI_IC_SECRET=$(jq -r '.data.attributes.incomingSecret' <<<"$result")
EI_CI_ACCESSKEY=$(jq -r '.data.attributes.outgoingToken' <<<"$result")
EI_CI_SECRET=$(jq -r '.data.attributes.outgoingSecret' <<<"$result")

{
  echo "EI_CI_ACCESSKEY=$EI_CI_ACCESSKEY"
  echo "EI_CI_SECRET=$EI_CI_SECRET"
  echo "EI_IC_ACCESSKEY=$EI_IC_ACCESSKEY"
  echo "EI_IC_SECRET=$EI_IC_SECRET"
} >"external_initiator_$1.env"

echo "EI has been added to Chainlink node"
echo "Done adding EI '$1'"
