# wander

<img src="./img/logo.png" width="35%"/>
<p>
    <a href="https://github.com/robinovitch61/wander/releases"><img src="https://shields.io/github/v/release/robinovitch61/wander.svg" alt="Latest Release"></a>
    <a href="https://pkg.go.dev/github.com/robinovitch61/wander?tab=doc"><img src="https://godoc.org/github.com/golang/gddo?status.svg" alt="GoDoc"></a>
    <a href="https://github.com/robinovitch61/wander/actions"><img src="https://github.com/robinovitch61/wander/workflows/build/badge.svg" alt="Build Status"></a>
</p>

An efficient terminal application/TUI for your [HashiCorp Nomad](https://www.nomadproject.io/) cluster.

- Browse jobs, allocations, tasks, and logs
- Exec to run commands in running tasks
- Tail global or targeted events
- Save any view as a local file
- See full specs

<div align="center">
   <img src="./img/jobs.png" width="100%"/>
   <img src="./img/allocations.png" width="49%"/>
   <img src="./img/logs.png" width="49%"/>
   <img src="./img/exec.png" width="49%"/>
   <img src="./img/events.png" width="49%"/>
   <img src="./img/save.png" width="49%"/>
   <img src="./img/spec.png" width="49%"/>
</div>
&nbsp;

`wander` is written with tools from [Charm](https://charm.sh/).

[Feature requests and bug reports for wander are welcome](https://github.com/robinovitch61/wander/issues/new/choose).

## Installation

### > Using homebrew

```shell
brew install robinovitch61/tap/wander
```

### > Download from GitHub

Download the relevant binary for your operating system (macOS = Darwin) from
the [latest Github release](https://github.com/robinovitch61/wander/releases). Unpack it, then move the binary to
somewhere accessible in your `PATH`, e.g. `mv ./wander /usr/local/bin`.

### > Using [go installed on your machine](https://go.dev/doc/install)

```shell
go install github.com/robinovitch61/wander@latest
```

### > Build from source

Clone this repo, build from source with `cd <cloned_repo> && go build`,
then move the binary to somewhere accessible in your `PATH`, e.g. `mv ./wander /usr/local/bin`.

## Usage

Run the app by running `wander` in a terminal. See `wander --help` and config section below for details.

## Configuration

`wander` can be configured in three ways:

1. Command line arguments, visible by running `wander --help`.
2. Environment variables. These map to the configuration file below (e.g. `nomad_addr` in yaml is the `NOMAD_ADDR`
   environment variable).
3. A yaml config file at `$HOME/.wander.yaml`, or a custom config file path passed to the `--config` argument. Complete
   example below.

Priority in order of highest to lowest is command line arguments, then environment variables, then the config file.

Example yaml file showing all options:

```shell
# Nomad address. Default "http://localhost:4646"
nomad_addr: http://localhost:4646

# Nomad token. Default ""
nomad_token: my-token

# Nomad region. Default ""
nomad_region: west

# Nomad namespace. Default "*"
nomad_namespace: "my-namespace"

# Nomad http auth, in the form of "user" or "user:pass". Default ""
nomad_http_auth: "username:password"

# Path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate. Default ""
nomad_cacert: "/path/to/cert"

# Path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate. If both cacert and capath are specified, cacert is used. Default ""
nomad_capath: "/path/to/cert/directory"

# Path to a PEM encoded client cert for TLS authentication to the Nomad server. Must also specify client key. Default ""
nomad_client_cert: "/path/to/cert"

# Path to an unencrypted PEM encoded private key matching the client cert. Default ""
nomad_client_key: "/path/to/key"

# The server name to use as the SNI host when connecting via TLS. Default ""
nomad_tls_server_name: server-name

# If "true", do not verify TLS certificates. Default "false"
nomad_skip_verify: true

# Seconds between updates for job & allocation pages. Disable with "-1". Default "2"
wander_update_seconds: 1

# Log byte offset from which logs start. Default "1000000"
wander_log_offset: 1000000

# If "true", copy the full path to file after save. Default "false"
wander_copy_save_path: true

# Topics to follow in event streams, comma-separated. Default "Job,Allocation,Deployment,Evaluation"
# see https://www.nomadproject.io/api-docs/events#event-stream
wander_event_topics: Job:my-job,Job:my-other-job,Allocation:my-job,Evaluation,Deployment:*

# Namespace used in stream for all events. "*" for all namespaces. Default "default"
wander_event_namespace: "*" # * needs surrounding "" in yaml

# jq (https://stedolan.github.io/jq/) query used for parsing events. "." to show entire event JSON. Default is:
# wander_event_jq_query: >
#   .Events[] | {
#      "1:Index": .Index,
#      "2:Topic": .Topic,
#      "3:Type": .Type,
#      "4:Name": .Payload | (.Job // .Allocation // .Deployment // .Evaluation) | (.JobID // .ID),
#      "5:AllocID": .Payload | (.Allocation // .Deployment // .Evaluation).ID[:8]
#   }
# The numbering exists to preserve ordering, as https://github.com/itchyny/gojq does not keep the order of object keys
wander_event_jq_query: .

# For `wander serve`. Hostname of the machine hosting the ssh server. Default "localhost"
wander_host: localhost

# For `wander serve`. Port for the ssh server. Default "21324"
wander_port: 21324

# For `wander serve`. Host key path for wander ssh server. Default none, i.e. ""
wander_host_key_path: .ssh/term_info_ed25519

# For `wander serve`. Host key PEM block for wander ssh server. Default none, i.e. ""
wander_host_key_pem: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaD1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAACFwAAAAdzc2gtcm
    ...
    XBMuWaiQMCZjAwAAAAp3YW5kZXItc3NoAQIEBAUGBw==
    -----END OPENSSH PRIVATE KEY-----

# Custom colors
wander_logo_color: "#DBBD70"
```

## SSH App

`wander` can be served via ssh application. For example, you could host an internal ssh application for your company
such that anyone on the internal network can `ssh -p <your-port> <your-host>` and immediately access `wander` without
installing or configuring anything.

Optionally, users can pass in their own nomad token with `ssh -p <port> <host> -t <token>`. The `-t` argument does not
stand for
token - it forces `ssh` to allocate a pty.

Serve the ssh app with `wander serve`.

## Trying It Out

You can try `wander` out by running a local nomad cluster in dev mode
following [these instructions](https://learn.hashicorp.com/tutorials/nomad/get-started-run?in=nomad/get-started):

```sh
# in first terminal session, start and leave nomad running in dev mode
sudo nomad agent -dev -bind 0.0.0.0 -log-level INFO

# in a different terminal session, create example job and run it
nomad job init
nomad job run example.nomad

# run wander
wander
```

## Development

The `scripts/dev.sh` script watches the source code and rebuilds the app on changes
using [entr](https://github.com/eradman/entr).

`wander` runs the built app. You must rerun it on rebuild.

If the `WANDER_DEBUG` environment variable is set to `true`, the `dev.Debug(s string)` function outputs logs
to `wander.log`.
