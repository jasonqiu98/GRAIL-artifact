# jepsen.arangodb.single

A Clojure library designed to test ArangoDB.

This project is also the generator of register histories for [`GRAIL-artifact`](../). Note that at line 115 of [`rw_register.clj`](./src/jepsen/arangodb/tests/rw_register.clj), we have increased the WAL chunk size in order to collect a full WAL log.

The bash script [`collect.sh`](./collect.sh) shows an example to collect the benchmark histories from this project, after following the instructions below to set up the environment.

## Set up your environment for this project

Here I use an example to set up the environment on Linux Mint 21 / Ubuntu 22.04.

1. Install [Virtualbox](https://www.virtualbox.org/wiki/Linux_Downloads) and [Vagrant](https://www.vagrantup.com/downloads) on your local machine.
2. Set up your local machine as the control node of this Jepsen test.
    - Install [Leiningen](https://leiningen.org/) with `sudo apt install leiningen`
    - Install other dependencies using `sudo apt install libjna-java gnuplot graphviz`
3. Create a public/private key pair for `vagrant` VMs
    1. Generate by `ssh-keygen`. Enter the path you want to save the key. Here I use the default path `~/.ssh/id_rsa`.
    2. Enter passphrase. Here I use the default value (empty for no passphrase). After that, enter same passphrase again.
    3. Then check the results. Here I have the following results.
    ```
    Your identification has been saved in ~/.ssh/id_rsa
    Your public key has been saved in ~/.ssh/id_rsa.pub
    The key fingerprint is:
    SHA256:vKQ4S16rSqkS7zCFShzXMNNVlog/8iRuPT7OWA+TjQI grail@grail-Legion-5-Pro-16ACH6H
    ```
    4. Copy, paste and overwrite the keys under the root folder of this project. Remember to replace the key path to your own path.
    ```
    cp ~/.ssh/id_rsa ./vagrant/vagrant_ssh_key
    cp ~/.ssh/id_rsa.pub ./vagrant/shared/vagrant_ssh_key.pub
    ```
4. Initialize Vagrant VMs
    1. Enter the `vagrant` folder by `cd vagrant`.
    2. Run `/bin/bash ./init/init.sh` in the project folder. This script destroys all the vagrant VMs and resets/starts new VMs. 
    3. Run `/bin/bash ./init/follow-up.sh` in the project folder. This script searches public keys of the VMs and writes to `known_hosts`.
        - Note! This script may fail as the VMs need time to get ready for further operations. You may need to wait a while before running this script.
        - A failure example (as you may see this first)
            - ```
              192.168.56.101 error
              192.168.56.102 error
              192.168.56.103 error
              192.168.56.104 error
              192.168.56.105 error
              ```
        - A success example
            - ```
              # 192.168.56.101:22 SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u1
              192.168.56.101 ok
              # 192.168.56.102:22 SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u1
              192.168.56.102 ok
              # 192.168.56.103:22 SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u1
              192.168.56.103 ok
              # 192.168.56.104:22 SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u1
              192.168.56.104 ok
              # 192.168.56.105:22 SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u1
              192.168.56.105 ok
              ```
    4. Return to the parent folder by `cd ..`.

Now the VM is started and waits for the following commands.

## Start the test

### Start the RW-register test

- Basic test (without partition): `/bin/bash ./run.sh --skip-vagrant --test-type rw`
- With time limit: `/bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 20`
- Increase number of concurrecy: `/bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 20 --concurrency 20`
- Other arguments:
  - `-r 10` (rate of operations)
  - generator-related args (See [link](https://github.com/jepsen-io/elle/blob/main/src/elle/list_append.clj#L920))
    - `--key-count` (number of distinct keys at any point)
    - `--min-txn-length` (minimum number of operations per txn)
    - `--max-txn-length` (maximum number of operations per txn)
    - `--max-writes-per-key` (maximum number of operations per key)
  - e.g. `/bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 20 -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3`

### And more

Click [here](doc/intro.md) to see the doc of the test(s).
