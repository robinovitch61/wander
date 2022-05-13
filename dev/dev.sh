#!/usr/bin/env zsh

ls **/*.go | entr -sr 'rm -f /usr/local/bin/wander && echo "building" && go build && cp wander /usr/local/bin && echo "built"'

