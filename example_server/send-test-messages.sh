#!/bin/bash -ex
PORT=${1:-1514}
echo "The example server should be started already, listening on UDP $PORT."

logger --udp --port $PORT --priority user.info Message one
