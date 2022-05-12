# wander

Wander is a terminal application for [Nomad by HashiCorp](https://www.nomadproject.io/).

It currently supports viewing jobs, allocations, tasks, and logs across a Nomad cluster.

It is written with the [Bubble Tea TUI framework from Charm](https://github.com/charmbracelet/bubbletea).

`wander` is in active development. Expect near term improvements. Feature requests in the form of issues are welcome.

## Installation

Currently, the best way to install `wander` is to clone this repo, build from source with `cd <cloned_repo> && go build`, then move the binary to somewhere accessible in your `PATH`, e.g. `mv ./wander /usr/local/bin`.

## Usage

`wander` requires two environment variables set:
- `NOMAD_ADDR`: path to nomad cluster
- `NOMAD_TOKEN`: token for auth against the HTTP API

You can try `wander` out by running a local nomad cluster in dev mode following [these instructions](https://learn.hashicorp.com/tutorials/nomad/get-started-run?in=nomad/get-started):
```sh
# in first terminal session, start and leave nomad running in dev mode
sudo nomad agent -dev -bind 0.0.0.0 -log-level INFO

# in a different terminal session, create example job and run it
nomad job init
nomad job run example.nomad

# run wander
NOMAD_ADDR=http://localhost:4646 NOMAD_TOKEN="blank" wander
```

## Development

The `dev/dev.sh` script watches the source code and rebuilds the app on changes using [entr](https://github.com/eradman/entr).

Run `./wander` to run the built app.

The `dev.Debug(s string)` function prints to `debug.log`, which you can tail with `tail -f debug.log`
