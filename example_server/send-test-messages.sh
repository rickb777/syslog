#!/bin/bash -ex

HOST=${1:-localhost}
PORT=${2:-1514}

# Before running this script, the example server should be running already on $HOST, listening on UDP $PORT.
#
# Try this out using
#   go build -o syslog ./example_server && ./syslog -port 1514 -file output.log
#
# For cross-compilation, the usual environment variables work fine, e.g.
#   GOOS=linux GOARCH=arm64 go build -o syslog.arm64 ./example_server

OPTS="--server $HOST --port $PORT -t taggy --stderr"
# the obsolete RFC3164 message format (still widely used)
logger $OPTS    --rfc3164        --priority user.notice  Message RFC3164 one
logger $OPTS -i --rfc3164        --priority user.notice  Message RFC3164 two
# the modern RFC5424 message format
logger $OPTS -i --rfc5424        --priority user.warning Message RFC5424 three
logger $OPTS -i --rfc5424=notq   --priority user.err     Message RFC5424 four
logger $OPTS -i --rfc5424=notime --priority user.info    Message RFC5424 five
logger $OPTS -i --rfc5424=nohost --priority user.info    Message RFC5424 six
logger $OPTS    --rfc5424        --priority user.debug   Message RFC5424 seven

