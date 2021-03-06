#!/bin/bash

# USAGE: ./e2e_multinode.sh COSMOS_NODE_BIN

set -e

# set ulimit early
ulimit -n 65536

CWD=$(pwd)

# These options can be overridden by env
CHAIN_ID="${CHAIN_ID:-injective-1}"
CHAIN_DIR="${CHAIN_DIR:-$CWD/var/data}"
DENOM="${DENOM:-inj}"

TEST_USDC_DENOM="${TEST_USDC_DENOM:-peggy0xc83DCEA3Ec44b7D3Ec70690BAb1e6292A80e6DC3}"
TEST_USDC_DECIMALS="${TEST_USDC_DECIMALS:-000000}"

TEST_LINK_DENOM="${TEST_LINK_DENOM:-peggy0x514910771AF9Ca656af840dff83E8264EcF986CA}"
TEST_LINK_DECIMALS="${TEST_LINK_DECIMALS:-000000000000000000}"

STAKE_DENOM="${STAKE_DENOM:-$DENOM}"
CLEANUP="${CLEANUP:-0}"
LOG_LEVEL=info
SCALE_FACTOR="${SCALE_FACTOR:-000000000000000000}"

# Default 3 account keys + 1 user key with no special grants
VAL0_KEY="val"
VAL0_MNEMONIC="remember huge castle bottom apology smooth avocado ceiling tent brief detect poem"
VAL1_KEY="val"
VAL1_MNEMONIC="capable dismiss rice income open wage unveil left veteran treat vast brave"
VAL2_KEY="val"
VAL2_MNEMONIC="jealous wrist abstract enter erupt hunt victory interest aim defy camp hair"

USER0_KEY="user0"
USER0_MNEMONIC="divide report just assist salad peanut depart song voice decide fringe stumble"
USER1_KEY="user1"
USER1_MNEMONIC="physical page glare junk return scale subject river token door mirror title"

ORACLE0_KEY="oracle0"
ORACLE0_MNEMONIC="bullet primary spider betray doctor truly cigar bulb whale bargain fence marble"
ORACLE1_KEY="oracle1"
ORACLE1_MNEMONIC="always impulse hobby nasty width find canyon grant juice doll scout inherit"
ORACLE2_KEY="oracle2"
ORACLE2_MNEMONIC="gloom kick buffalo long cruel refuse bind rather quiz chicken deer sausage"
ORACLE3_KEY="oracle3"
ORACLE3_MNEMONIC="science rabbit damp acquire clock oven february heavy path meat act essence"

NEWLINE=$'\n'

hdir="$CHAIN_DIR/$CHAIN_ID"

if [[ $# -eq 0 ]]; then
	echo "Usage: $0 COSMOS_NODE_BIN"
	exit 1
fi

if ! command -v jq &> /dev/null
then
    echo "⚠️ jq command could not be found!"
    echo "jq is a lightweight and flexible command-line JSON processor."
    echo "Install it by checking https://stedolan.github.io/jq/download/"
    exit 1
fi

# Expect Chain ID to be provided
if [[ -z "$CHAIN_ID" ]]; then
  echo "Please provide Cosmos CHAIN_ID env"
  exit 1
fi

# Expect data prefix to be provided
if [[ -z "$CHAIN_DIR" ]]; then
  echo "Please provide CHAIN_DIR data prefix"
  exit 1
fi

NODE_BIN="$1"

echo "Using $CHAIN_ID as Chain ID and $CHAIN_DIR as data prefix."
echo "Using $DENOM as Cosmos Coin Denom."
if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
	echo "Will remove $CHAIN_DIR"
fi
echo "Press ^C if you don't agree.."

killall "$NODE_BIN" &>/dev/null || true

sleep 3

if [[ "$CLEANUP" == 1 || "$CLEANUP" == "1" ]]; then
	rm -rf "$CHAIN_DIR"
fi

# Folders for nodes
n0dir="$hdir/n0"
n1dir="$hdir/n1"
n2dir="$hdir/n2"

# Home flag for folder
home0="--home $n0dir"
home1="--home $n1dir"
home2="--home $n2dir"

# Config directories for nodes
n0cfgDir="$n0dir/config"
n1cfgDir="$n1dir/config"
n2cfgDir="$n2dir/config"

# Config files for nodes
n0cfg="$n0cfgDir/config.toml"
n1cfg="$n1cfgDir/config.toml"
n2cfg="$n2cfgDir/config.toml"

# App config files for nodes
n0app="$n0cfgDir/app.toml"
n1app="$n1cfgDir/app.toml"
n2app="$n2cfgDir/app.toml"

# Common flags
kbt="--keyring-backend test"
cid="--chain-id $CHAIN_ID"

PASSPHRASE="12345678"

# Check if the data dir has been initialized already
if [[ ! -d "$hdir" ]]; then
	echo "Creating 3x $NODE_BIN validators with chain-id=$CHAIN_ID"

	# Build genesis file and create accounts
	if [[ "$STAKE_DENOM" != "$DENOM" ]]; then
		coins="1000000$SCALE_FACTOR$STAKE_DENOM,1000000$SCALE_FACTOR$DENOM"
	else
		coins="1000000$SCALE_FACTOR$DENOM"
	fi

	coins_user="10000000$SCALE_FACTOR$DENOM,10000000000$TEST_USDC_DECIMALS$TEST_USDC_DENOM,10000000$TEST_LINK_DECIMALS$TEST_LINK_DENOM"
	coins_oracle="10000000$SCALE_FACTOR$DENOM" # only for gas

	echo "initializing node homes..."

	# Initialize the home directories of each node
	$NODE_BIN $home0 $cid init n0 &>/dev/null
	$NODE_BIN $home1 $cid init n1 &>/dev/null
	$NODE_BIN $home2 $cid init n2 &>/dev/null

	# Generate new random keys
	# $NODE_BIN $home0 keys add val $kbt &>/dev/null
	# $NODE_BIN $home1 keys add val $kbt &>/dev/null
	# $NODE_BIN $home2 keys add val $kbt &>/dev/null

	# Import keys from mnemonics
	yes "$VAL0_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $VAL0_KEY $kbt --recover
	yes "$VAL1_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $VAL1_KEY $kbt --recover
	yes "$VAL2_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $VAL2_KEY $kbt --recover

	yes "$USER0_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $USER0_KEY $kbt --recover
	yes "$USER0_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $USER0_KEY $kbt --recover &>/dev/null
	yes "$USER0_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $USER0_KEY $kbt --recover &>/dev/null

	yes "$USER1_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $USER1_KEY $kbt --recover
	yes "$USER1_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $USER1_KEY $kbt --recover &>/dev/null
	yes "$USER1_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $USER1_KEY $kbt --recover &>/dev/null

	yes "$ORACLE0_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $ORACLE0_KEY $kbt --recover
	yes "$ORACLE0_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $ORACLE0_KEY $kbt --recover &>/dev/null
	yes "$ORACLE0_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $ORACLE0_KEY $kbt --recover &>/dev/null

	yes "$ORACLE1_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $ORACLE1_KEY $kbt --recover
	yes "$ORACLE1_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $ORACLE1_KEY $kbt --recover &>/dev/null
	yes "$ORACLE1_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $ORACLE1_KEY $kbt --recover &>/dev/null

	yes "$ORACLE2_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $ORACLE2_KEY $kbt --recover
	yes "$ORACLE2_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $ORACLE2_KEY $kbt --recover &>/dev/null
	yes "$ORACLE2_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $ORACLE2_KEY $kbt --recover &>/dev/null

	yes "$ORACLE3_MNEMONIC$NEWLINE" | $NODE_BIN $home0 keys add $ORACLE3_KEY $kbt --recover
	yes "$ORACLE3_MNEMONIC$NEWLINE" | $NODE_BIN $home1 keys add $ORACLE3_KEY $kbt --recover &>/dev/null
	yes "$ORACLE3_MNEMONIC$NEWLINE" | $NODE_BIN $home2 keys add $ORACLE3_KEY $kbt --recover &>/dev/null

	# Add addresses to genesis
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $VAL0_KEY -a $kbt) $coins &>/dev/null
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home1 keys show $VAL1_KEY -a $kbt) $coins &>/dev/null
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home2 keys show $VAL2_KEY -a $kbt) $coins &>/dev/null

	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $USER0_KEY -a $kbt) $coins_user &>/dev/null
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $USER1_KEY -a $kbt) $coins_user &>/dev/null

	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $ORACLE0_KEY -a $kbt) $coins_oracle &>/dev/null
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $ORACLE1_KEY -a $kbt) $coins_oracle &>/dev/null
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $ORACLE2_KEY -a $kbt) $coins_oracle &>/dev/null
	$NODE_BIN $home0 add-genesis-account $($NODE_BIN $home0 keys show $ORACLE3_KEY -a $kbt) $coins_oracle &>/dev/null

	# Patch genesis.json to better configure stuff for testing purposes
	if [[ "$STAKE_DENOM" == "$DENOM" ]]; then
		cat $n0cfgDir/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="'$DENOM'"' > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
		cat $n0cfgDir/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="'$DENOM'"' > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
		cat $n0cfgDir/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'$DENOM'"' > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
		cat $n0cfgDir/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="'$DENOM'"' > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
	fi

	echo "NOTE: Setting Governance Voting Period to 5 seconds for rapid testing"
	cat $n0cfgDir/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="5s"' > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json
	echo "NOTE: Setting OCR module payout block interval to 10 blocks for rapid testing"
	cat $n0cfgDir/genesis.json | jq '.app_state["ocr"]["params"]["payout_block_interval"]="10"' > $n0cfgDir/tmp_genesis.json && mv $n0cfgDir/tmp_genesis.json $n0cfgDir/genesis.json

	# Copy genesis around to sign
	cp $n0cfgDir/genesis.json $n1cfgDir/genesis.json
	cp $n0cfgDir/genesis.json $n2cfgDir/genesis.json

	# Create gentxs and collect them in n0
	$NODE_BIN $home0 gentx $VAL0_KEY 1000$SCALE_FACTOR$STAKE_DENOM $kbt $cid &>/dev/null
	$NODE_BIN $home1 gentx $VAL1_KEY 1000$SCALE_FACTOR$STAKE_DENOM $kbt $cid &>/dev/null
	$NODE_BIN $home2 gentx $VAL2_KEY 1000$SCALE_FACTOR$STAKE_DENOM $kbt $cid &>/dev/null

	cp $n1cfgDir/gentx/*.json $n0cfgDir/gentx/
	cp $n2cfgDir/gentx/*.json $n0cfgDir/gentx/
	$NODE_BIN $home0 collect-gentxs &>/dev/null

	# Copy genesis file into n1 and n2s
	cp $n0cfgDir/genesis.json $n1cfgDir/genesis.json
	cp $n0cfgDir/genesis.json $n2cfgDir/genesis.json

	# Run this to ensure everything worked and that the genesis file is setup correctly
	$NODE_BIN $home0 validate-genesis
	$NODE_BIN $home1 validate-genesis
	$NODE_BIN $home2 validate-genesis

	# Actually a cross-platform solution, sed is rubbish
	# Example usage: $REGEX_REPLACE 's/^param = ".*?"/param = "100"/' config.toml
	REGEX_REPLACE="perl -i -pe"

	echo "regex replacing config variables"

	$REGEX_REPLACE 's|addr_book_strict = true|addr_book_strict = false|g' $n0cfg
	$REGEX_REPLACE 's|external_address = ""|external_address = "tcp://127.0.0.1:26657"|g' $n0cfg
	$REGEX_REPLACE 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26657"|g' $n0cfg
	$REGEX_REPLACE 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $n0cfg
	$REGEX_REPLACE 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $n0cfg
	$REGEX_REPLACE 's|timeout_commit = ".*?"|timeout_commit = "1s"|g' $n0cfg

	$REGEX_REPLACE 's|addr_book_strict = true|addr_book_strict = false|g' $n1cfg
	$REGEX_REPLACE 's|external_address = ""|external_address = "tcp://127.0.0.1:26667"|g' $n1cfg
	$REGEX_REPLACE 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26667"|g' $n1cfg
	$REGEX_REPLACE 's|"tcp://0.0.0.0:26656"|"tcp://0.0.0.0:26666"|g' $n1cfg
	$REGEX_REPLACE 's|"localhost:6060"|"localhost:6061"|g' $n1cfg
	$REGEX_REPLACE 's|"tcp://0.0.0.0:10337"|"tcp://0.0.0.0:11337"|g' $n1app
	$REGEX_REPLACE 's|"0.0.0.0:1317"|"0.0.0.0:1417"|g' $n1app
	$REGEX_REPLACE 's|"0.0.0.0:9900"|"0.0.0.0:9901"|g' $n1app
	$REGEX_REPLACE 's|"0.0.0.0:9091"|"0.0.0.0:9092"|g' $n1app
	$REGEX_REPLACE 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $n1cfg
	$REGEX_REPLACE 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $n1cfg
	$REGEX_REPLACE 's|timeout_commit = ".*?"|timeout_commit = "1s"|g' $n1cfg

	$REGEX_REPLACE 's|addr_book_strict = true|addr_book_strict = false|g' $n2cfg
	$REGEX_REPLACE 's|external_address = ""|external_address = "tcp://127.0.0.1:26677"|g' $n2cfg
	$REGEX_REPLACE 's|"tcp://127.0.0.1:26657"|"tcp://0.0.0.0:26677"|g' $n2cfg
	$REGEX_REPLACE 's|"tcp://0.0.0.0:26656"|"tcp://0.0.0.0:26676"|g' $n2cfg
	$REGEX_REPLACE 's|"localhost:6060"|"localhost:6062"|g' $n2cfg
	$REGEX_REPLACE 's|"tcp://0.0.0.0:10337"|"tcp://0.0.0.0:12337"|g' $n2app
	$REGEX_REPLACE 's|"0.0.0.0:1317"|"0.0.0.0:1517"|g' $n2app
	$REGEX_REPLACE 's|"0.0.0.0:9900"|"0.0.0.0:9902"|g' $n2app
	$REGEX_REPLACE 's|"0.0.0.0:9091"|"0.0.0.0:9093"|g' $n2app
	$REGEX_REPLACE 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $n2cfg
	$REGEX_REPLACE 's|log_level = "info"|log_level = "'$LOG_LEVEL'"|g' $n2cfg
	$REGEX_REPLACE 's|timeout_commit = ".*?"|timeout_commit = "1s"|g' $n2cfg

	# Set peers for all three nodes
	peer0="$($NODE_BIN $home0 tendermint show-node-id)\@127.0.0.1:26656"
	peer1="$($NODE_BIN $home1 tendermint show-node-id)\@127.0.0.1:26666"
	peer2="$($NODE_BIN $home2 tendermint show-node-id)\@127.0.0.1:26676"
	$REGEX_REPLACE 's|persistent_peers = ""|persistent_peers = "'$peer1','$peer2'"|g' $n0cfg
	$REGEX_REPLACE 's|persistent_peers = ""|persistent_peers = "'$peer0','$peer2'"|g' $n1cfg
	$REGEX_REPLACE 's|persistent_peers = ""|persistent_peers = "'$peer0','$peer1'"|g' $n2cfg
fi # data dir check

# Start the instances
echo "Starting nodes..."

$NODE_BIN $home0 start --grpc.address="0.0.0.0:9900" > $hdir.n0.log 2>&1 &
$NODE_BIN $home1 start --grpc.address="0.0.0.0:9901" > $hdir.n1.log 2>&1 &
$NODE_BIN $home2 start --grpc.address="0.0.0.0:9902" > $hdir.n2.log 2>&1 &

# Wait for chains to start
echo "Waiting for chains to start..."
sleep 8

echo
echo "Logs:"
echo "  * tail -f ./var/data/$CHAIN_ID.n0.log"
echo "  * tail -f ./var/data/$CHAIN_ID.n1.log"
echo "  * tail -f ./var/data/$CHAIN_ID.n2.log"
echo 
echo "Env for easy access:"
echo "export H1='--home ./var/data/$CHAIN_ID/n0/'"
echo "export H2='--home ./var/data/$CHAIN_ID/n1/'"
echo "export H3='--home ./var/data/$CHAIN_ID/n2/'"
echo 
echo "Command Line Access:"
echo "  * $NODE_BIN --home ./var/data/$CHAIN_ID/n0 status"
echo "  * $NODE_BIN --home ./var/data/$CHAIN_ID/n1 status"
echo "  * $NODE_BIN --home ./var/data/$CHAIN_ID/n2 status"
