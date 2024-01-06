#!/usr/bin/env zsh

# https://github.com/charmbracelet/vhs
# expects nomad to already be running locally
# should be run from scripts directory

rm -f /tmp/my_logs.txt
latest_tag=$(git describe --tags --abbrev=0)
tag=${1:-$latest_tag}
./rebuild.sh "$tag"
mv ~/.wander.{yaml,yaml.tmp}
vhs vhs.tape
open ../img/wander.gif
mv ~/.wander.{yaml.tmp,yaml}
rm -f my_logs.txt
rm -f ../img/screenshots/README.md
for file in ../img/screenshots/*.png; do
  filename=$(basename -- "$file")
  filename="${filename%.*}"
  echo "# ${filename//_/ }" >> ../img/screenshots/README.md
  echo "![](./$filename.png)" >> ../img/screenshots/README.md
done
