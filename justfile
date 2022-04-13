set shell := ["zsh", "-cu"]

dev:
  ls **/*.go | entr -sr 'rm -f wander && echo "building" && go build && echo "built"'
