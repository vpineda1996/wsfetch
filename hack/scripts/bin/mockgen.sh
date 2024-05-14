#!/bin/bash 

# Will create a folder called mocks where mocks 
# will be stored
# Usage:
# mockgen.sh <file with interface> <interfaces to mock>

set -xeu

if [ $# -lt 2 ]; then
    echo "Please refer to the right usage of this command line"
    echo "   mockgen . Conn"
    exit 1
fi

dstDir=$(pwd)/mocks
mkdir -p ${dstDir}

mockgen -package mocks -destination $(pwd)/mocks/$1 -source "$1" "${@:2}"