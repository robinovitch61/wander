#!/usr/bin/env zsh

DEV_CMD=$(cat <<-EOM
rm -f $GOPATH/bin/wander-ssh && \
echo "building $(date +"%T")" && \
go build -o wander-ssh ./cmd/wander-ssh && \
mv wander-ssh $GOPATH/bin && \
echo "built"
EOM
)

ls **/*.go | entr -sr "${DEV_CMD}"
