#1/usr/bin/env sh

set -e

nomad job init
sed -i "" -E "s/job \"example\"/job \"${1:-example}\"/" example.nomad
nomad job run -detach example.nomad
rm ./example.nomad
