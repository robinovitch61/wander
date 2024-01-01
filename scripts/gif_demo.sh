#!/usr/bin/env zsh

# https://github.com/charmbracelet/vhs
# expects nomad to already be running locally
# should be run from scripts directory

# get the latest tag on this branch
latest_tag=$(git describe --tags --abbrev=0)
./rebuild.sh "$latest_tag"
mv ~/.wander.{yaml,yaml.tmp}
vhs vhs.tape
mv ~/.wander.{yaml.tmp,yaml}
rm -f my_logs.txt
rm -f ../img/screenshots/README.md
for file in ../img/screenshots/*.png; do
  filename=$(basename -- "$file")
  filename="${filename%.*}"
  echo "# ${filename//_/ }" >> ../img/screenshots/README.md
  echo "![](./$filename.png)" >> ../img/screenshots/README.md
done
