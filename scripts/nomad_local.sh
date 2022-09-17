#1/usr/bin/env sh

set -e

current_nomad_pid=$(lsof | rg ^nomad | head -n 1 | awk '{ print $2 }')

if [ ! -z $current_nomad_pid ]; then
  kill -9 $current_nomad_pid
fi

nomad agent -dev -bind 0.0.0.0 -log-level INFO
