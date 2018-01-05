#!/bin/sh
rm -rf /Users/rolong/blockchain/projects/etz_data/*
./build/bin/getz --datadir=/Users/rolong/blockchain/projects/etz_data/ init ./my-genesis.json
./build/bin/getz --datadir=/Users/rolong/blockchain/projects/etz_data/ --password ./password account new
./build/bin/getz --datadir=/Users/rolong/blockchain/projects/etz_data/ --password ./password account new
./build/bin/getz --datadir=/Users/rolong/blockchain/projects/etz_data/ --password ./password account new
