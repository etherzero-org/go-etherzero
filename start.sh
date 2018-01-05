#!/bin/bash
./build/bin/getz --datadir=/Users/rolong/blockchain/projects/etz_data/ \
    --rpc --rpcaddr="0.0.0.0" --rpccorsdomain="*" --unlock '0' \
    --password ./password   --nodiscover --maxpeers '5' --networkid '2017' \
    console
