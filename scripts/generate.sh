#1/usr/bin/env sh

set -e

run_job() {
    nomad job init
    sed -i "" -E "s/job \"example\"/job \"${1:-example}\"/" example.nomad
    nomad job run -detach example.nomad
    rm ./example.nomad
}

run_job alright_stop
run_job collaborate_and_listen
run_job ice_is_back
run_job with_a_brand_new_invention
