#1/usr/bin/env sh

set -e

nomad namespace apply -description "My super cool namespace" my-namespace
rm -f example.nomad
nomad job init -short
sed -i "" -E '1a\
  namespace = "default"
' example.nomad

run_job() {
    sed -i "" -E "s/job \".*\"/job \"${1:-example}\"/" example.nomad
    sed -i "" -E "s/namespace = \".*\"/namespace = \"${2:-default}\"/" example.nomad
    cat example.nomad
    nomad job run -detach example.nomad
}

run_job alright_stop
run_job collaborate_and_listen
run_job ice_is_back my-namespace
run_job with_a_brand_new_invention

rm example.nomad
