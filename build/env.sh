#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
etzdir="$workspace/src/github.com/ethzero"
if [ ! -L "$etzdir/go-ethzero" ]; then
    mkdir -p "$etzdir"
    cd "$etzdir"
    ln -s ../../../../../. go-ethzero
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$etzdir/go-ethzero"
PWD="$etzdir/go-ethzero"

# Launch the arguments with the configured environment.
exec "$@"
