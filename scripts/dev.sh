#!/usr/bin/env zsh

DEV_CMD=$(cat <<-EOM
rm -f $GOPATH/bin/wander && \
echo "building $(date +"%T")" && \
go build -o wander ./cmd/wander && \
mv wander $GOPATH/bin && \
echo "built"
EOM
)

ls **/*.go | entr -sr "${DEV_CMD}"
