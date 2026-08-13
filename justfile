ssh_opts := "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"

# List available recipes
default:
    @just --list

# Build the binary
build:
    go build -o sshblog .

# Build and run the server
run: build
    ./sshblog

# Run the server in the background and ssh in; clean up on exit
dev:
    #!/usr/bin/env bash
    set -euo pipefail
    go build -o sshblog .
    ./sshblog > /tmp/sshblog.log 2>&1 &
    server_pid=$!
    trap 'kill "$server_pid" 2>/dev/null || true' EXIT
    sleep 0.5
    ssh -tt localhost {{ssh_opts}}

# Connect to a running server over ssh
connect:
    ssh localhost {{ssh_opts}}

# Remove build artifacts
clean:
    rm -f sshblog
