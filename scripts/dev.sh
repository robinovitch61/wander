#!/usr/bin/env zsh

ls **/*.go | entr -sr 'rm -f $GOPATH/bin/wander && echo "building $(date +"%T")" && go build && mv wander $GOPATH/bin && echo "built"'

