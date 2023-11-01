#!/usr/bin/env bash

LOOT_SURVIVOR_BINARY=${LOOT_SURVIVOR_BINARY:-"loot-survivor"}
DATA_DIR=${DATA_DIR:-"$(pwd)/data"}
CRAWL_INTERVAL=${CRAWL_INTERVAL:-1800}

set -e

mkdir -p "$DATA_DIR"

if [ -z "$LOOT_SURVIVOR_CONTRACT_ADDRESS" ]
then
    echo "LOOT_SURVIVOR_CONTRACT_ADDRESS is not set"
    exit 1
fi

if [ ! -f "$DATA_DIR/last_crawled_block.txt" ]
then
    echo "-1" > "$DATA_DIR/last_crawled_block.txt"
fi

LAST_CRAWLED_BLOCK=$(cat "$DATA_DIR/last_crawled_block.txt")
NEXT_BLOCK=$((LAST_CRAWLED_BLOCK + 1))

CURRENT_BLOCK=$($LOOT_SURVIVOR_BINARY stark block-number)

if [ "$CURRENT_BLOCK" -le "$NEXT_BLOCK" ]
then
    echo "No new blocks to crawl"
    exit 0
fi

while true
do
    echo "Updating events for blocks ${NEXT_BLOCK}-${CURRENT_BLOCK}"
    $LOOT_SURVIVOR_BINARY stark events \
        -N 1000 \
        --confirmations 5 \
        --hot-interval 10 \
        --cold-interval 1000 \
        --contract "$LOOT_SURVIVOR_CONTRACT_ADDRESS" \
        --from "$NEXT_BLOCK" \
        --to "$CURRENT_BLOCK" \
        > "$DATA_DIR/events-$NEXT_BLOCK-$CURRENT_BLOCK.jsonl"

    echo "$CURRENT_BLOCK" >"$DATA_DIR/last_crawled_block.txt"

    echo "Sleeping for $CRAWL_INTERVAL seconds"
    sleep "$CRAWL_INTERVAL"
done
