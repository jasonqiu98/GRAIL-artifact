#!/bin/bash

# exit if any command fails
set -e

if [[ "$1" = "--skip-vagrant" ]]; then
    echo "skip vagrant up"
    MORE_ARGS="${@:2}"
else
    cd vagrant
    vagrant halt
    vagrant up
    cd ..
    MORE_ARGS="${@:1}"
fi

# Make sure our private key has correct permssions
chmod 600 ./vagrant/vagrant_ssh_key

lein run test \
    --nodes-file ./vagrant/nodes-vagrant.txt \
    --password \
    --ssh-private-key ./vagrant/vagrant_ssh_key \
    --username vagrant \
    $MORE_ARGS