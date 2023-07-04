#!/usr/bin/env zsh

# https://github.com/charmbracelet/vhs
# expects nomad to already be running locally
# should be run from scripts directory
mv ~/.wander.{yaml,yaml.tmp}
vhs vhs.tape
mv ~/.wander.{yaml.tmp,yaml}
rm my_logs.txt
open ../img/wander.gif
