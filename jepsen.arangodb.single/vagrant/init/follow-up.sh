#!/bin/bash

# search public keys of the VMs and write to known_hosts
for node in $(awk '{print}' < nodes-vagrant.txt); do
    ssh-keyscan -t rsa $node >> ~/.ssh/known_hosts
    if [[ "$?" = "0" ]]; then
        echo $node "ok"
    else
        echo $node "error"
    fi
done
