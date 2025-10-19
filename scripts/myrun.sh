#!/bin/bash
# myrun.sh
# Simple script re-build start 2 background processes:
# 1. tempest-homekit-go: console with [additional flags...]
# 2. tempest-homekit-go: editor
# Usage: ./scripts/myrun.sh [additional flags...]
# Example: ./scripts/myrun.sh --disable-homekit --web-port 8086
set -e

./scripts/kill-all.sh # Ensure no other instances are running

# Build and run
go build

./tempest-homekit-go --alarms @tempest-alarms.json "$@" &
./tempest-homekit-go --alarms-edit @tempest-alarms.json &
./tempest-homekit-go --webhook-listener-port 8082 &
