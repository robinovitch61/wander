#!/usr/bin/env zsh

# https://github.com/charmbracelet/vhs
# expects nomad to already be running locally
# should be run from scripts directory
./rebuild.sh "$1"
mv ~/.wander.{yaml,yaml.tmp}
vhs vhs.tape
mv ~/.wander.{yaml.tmp,yaml}
rm -f my_logs.txt
