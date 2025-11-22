#!/bin/bash
# Test that .env is loaded
export HISTORY_POINTS=12345
./tempest-homekit-go --use-generated-weather --disable-homekit --history-read 2>&1 | grep -i "generating.*historical" | head -1
