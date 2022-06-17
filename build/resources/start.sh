#!/bin/bash
set -e

#Launch the gateway
./apinto start

echo "APINTO start Success!" >> a.out
tail -f a.out