#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
intchaindir="$workspace/src/github.com/intfoundation"
if [ ! -L "$intchaindir/intchain" ]; then
    mkdir -p "$intchaindir"
    cd "$intchaindir"
    ln -s ../../../../../. intchain
    cd "$root"
fi

# Set up the environment to use the workspace.
#GOPATH="$workspace"

#export GOPATH

# Run the command inside the workspace.
cd "$intchaindir/intchain"
PWD="$intchaindir/intchain"

# build intchain client
go build -o $root/bin/intchain ./cmd/intchain/
