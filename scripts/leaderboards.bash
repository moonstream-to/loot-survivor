#!/usr/bin/env bash

LOOT_SURVIVOR_BINARY=${LOOT_SURVIVOR_BINARY:-"loot-survivor"}
PROJECT_ROOT_DIR="$(dirname "$0")/.."
DATA_DIR=${DATA_DIR:-"$PROJECT_ROOT_DIR/data"}
UPDATE_INTERVAL=${CRAWL_INTERVAL:-1800}

if [ -z "$MOONSTREAM_ACCESS_TOKEN" ]
then
    echo "MOONSTREAM_ACCESS_TOKEN is not set"
    exit 1
fi
if [ -z "$BEAST_SLAYERS_LEADERBOARD_ID" ]
then
    echo "BEAST_SLAYERS_LEADERBOARD_ID is not set"
    exit 1
fi

if [ -z "$ARTFUL_DODGERS_LEADERBOARD_ID" ]
then
    echo "ARTFUL_DODGERS_LEADERBOARD_ID is not set"
    exit 1
fi

set -e -o pipefail

while true
do
    echo "Updating Beast Slayers leaderboard"
    cat "$DATA_DIR"/events-* | "$LOOT_SURVIVOR_BINARY" leaderboards beast-slayers -i - -o "$DATA_DIR/slayers.json" --push --leaderboard-id "$BEAST_SLAYERS_LEADERBOARD_ID"

    echo "Updating Artful Dodgers leaderboard"
    cat "$DATA_DIR"/events-* | "$LOOT_SURVIVOR_BINARY" leaderboards artful-dodgers -i - -o "$DATA_DIR/dodgers.json" --push --leaderboard-id "$ARTFUL_DODGERS_LEADERBOARD_ID"

    echo "Sleeping for $UPDATE_INTERVAL seconds"
    sleep "$UPDATE_INTERVAL"
done
