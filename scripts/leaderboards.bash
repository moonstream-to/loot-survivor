#!/usr/bin/env bash

DATA_DIR=${DATA_DIR:-"$(pwd)/data"}

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

cat "$DATA_DIR"/events-* | ./loot-survivor leaderboards beast-slayers -i - -o "$DATA_DIR/slayers.json" --push --leaderboard-id "$BEAST_SLAYERS_LEADERBOARD_ID"

cat "$DATA_DIR"/events-* | ./loot-survivor leaderboards artful-dodgers -i - -o "$DATA_DIR/dodgers.json" --push --leaderboard-id "$ARTFUL_DODGERS_LEADERBOARD_ID"
