#!/usr/bin/env zsh

thisdir=${0:a:h}

ls $thisdir/../**/*.go | entr -src $thisdir/rebuild.sh
