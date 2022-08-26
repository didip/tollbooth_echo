#!/bin/bash

# set -x
LAUNCH_DIR=$(pwd); SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"; cd $SCRIPT_DIR; cd ..; SCRIPT_PARENT_DIR=$(pwd);

cd $SCRIPT_PARENT_DIR

URL="localhost:4444/"

for _ in {1..5}; do
  curl -s ${URL}
done

cd $LAUNCH_DIR