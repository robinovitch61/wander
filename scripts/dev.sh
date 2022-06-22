#!/usr/bin/env zsh

thisdir=${0:a:h}

DEV_CMD=$(cat <<-EOM
rm -f $GOPATH/bin/wander && \
echo "building $(date +"%T")" && \
go build $thisdir/.. && \
mv $thisdir/../wander $GOPATH/bin && \
echo "built"
EOM
)

ls $thisdir/../**/*.go | entr -sr "${DEV_CMD}"
