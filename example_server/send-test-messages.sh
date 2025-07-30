#!/bin/bash -ex
PORT=${1:-1514}
# The example server should be started already, listening on UDP $PORT.
# For example, use
#   go build -o syslog ./example_server && ./syslog -port 1514 -file output.log

logger --server localhost --port $PORT -t taggy -s    --rfc3164 --priority user.notice Message RFC3164 one
logger --server localhost --port $PORT -t taggy -s -i --rfc3164 --priority user.notice Message RFC3164 two
logger --server localhost --port $PORT -t taggy -s -i --rfc5424 --priority user.warning Message RFC5424 three
logger --server localhost --port $PORT -t taggy -s -i --rfc5424=notq --priority user.err Message RFC5424 four
logger --server localhost --port $PORT -t taggy -s -i --rfc5424=notime --priority user.info Message RFC5424 five
logger --server localhost --port $PORT -t taggy -s -i --rfc5424=nohost --priority user.info Message RFC5424 six
logger --server localhost --port $PORT -t taggy -s    --rfc5424 --priority user.debug Message RFC5424 seven