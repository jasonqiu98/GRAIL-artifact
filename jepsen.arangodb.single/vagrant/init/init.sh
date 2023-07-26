#!/bin/bash

# forcibly destroy all the VMs
vagrant destroy --force

# remove all the known_hosts records of the nodes that we're going to use
for node in $(awk '{print}' < nodes-vagrant.txt); do
    ssh-keygen -f ~/.ssh/known_hosts -R $node
done

# initialize all the VMs
vagrant up
