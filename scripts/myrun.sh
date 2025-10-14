#!/bin/bash
# myrun.sh
# Simple script to run multiple instances of tempest-homekit-go with different alarm files
# Usage: ./scripts/myrun.sh [additional flags...]
# Example: ./scripts/myrun.sh --disable-homekit --web-port 8086
./scripts/kill-all.sh # Ensure no other instances are running
./tempest-homekit-go --alarms @tempest-alarms.json "$@" &
./tempest-homekit-go --alarms-edit @tempest-alarms.json "$@" &
