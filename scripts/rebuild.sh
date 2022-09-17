#!/usr/bin/env zsh

thisdir=${0:a:h}
rm -f $GOPATH/bin/wander
echo "building $(date +"%T")"
go build -o $thisdir/wander $thisdir/..
if [ -f $thisdir/wander ]; then
  mv $thisdir/wander $GOPATH/bin/wander
  echo "built"
fi
