#!/usr/bin/env zsh

thisdir=${0:a:h}
rm -f $GOPATH/bin/wander
echo "building $(date +"%T")"
if [ $# -eq 1 ]; then
  go build -ldflags "-X github.com/robinovitch61/wander/cmd.Version=$1" -o $thisdir/wander $thisdir/..
else
  go build -o $thisdir/wander $thisdir/..
fi
if [ -f $thisdir/wander ]; then
  mv $thisdir/wander $GOPATH/bin/wander
  echo "built"
fi
