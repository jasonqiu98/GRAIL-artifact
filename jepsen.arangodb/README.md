# jepsen.arangodb

A Clojure library designed to test ArangoDB.

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

### Start the register test

- Basic test (without partition): `/bin/bash ./run.sh --skip-vagrant --test-type register`
  - The `--skip-vagrant` handle is recommended as it will let the program skip vagrant setup. If you have already followed the previous instructions, do include this handler; otherwise the VMs will be started by `vagrant up`, and you might lose some necessary configurations to run the program. Note that the `--skip-vagrant` handle should be placed immediately after `/bin/bash ./run.sh` as part of the start of the command.
- With partition: `/bin/bash ./run.sh --skip-vagrant --test-type register --nemesis-type partition`
- With partition and time limit: `/bin/bash ./run.sh --skip-vagrant --test-type register --time-limit 20 --nemesis-type partition`
- Increase number of concurrecy: `/bin/bash ./run.sh --skip-vagrant --test-type register --time-limit 20 --concurrency 20 --nemesis-type partition`
- Other arguments: `/bin/bash ./run.sh --skip-vagrant --test-type register --time-limit 20 -r 10 --concurrency 20 --ops-per-key 10 --threads-per-group 5 --nemesis-type noop`

Errors found?

- Consider generate a new public/private key pair by `ssh-keygen` and goes along the process again. Check your `ssh-agent` as well.
- If the nemesis type is `partition`, you might see some errors in the end of the output like the following one. This is an internal error raised because of the force shutdown of the program. See it more like a warning instead of an actual error :)

```
ERROR [2022-12-05 13:08:26,006] pool-6-thread-1 - com.arangodb.internal.velocystream.internal.MessageStore Reached the end of the stream.
java.io.IOException: Reached the end of the stream.
        at com.arangodb.internal.velocystream.internal.VstConnection.readBytesIntoBuffer(VstConnection.java:348)
        at com.arangodb.internal.velocystream.internal.VstConnection.readBytes(VstConnection.java:340)
        at com.arangodb.internal.velocystream.internal.VstConnection.readChunk(VstConnection.java:315)
        at com.arangodb.internal.velocystream.internal.VstConnection.lambda$open$0(VstConnection.java:212)
        at java.base/java.util.concurrent.FutureTask.run(FutureTask.java:264)
        at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1128)
        at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:628)
        at java.base/java.lang.Thread.run(Thread.java:829)
```

### Start the list append test

- Basic test (without partition): `/bin/bash ./run.sh --skip-vagrant --test-type la`
- With partition: `/bin/bash ./run.sh --skip-vagrant --test-type la --nemesis-type partition`
- With partition and time limit: `/bin/bash ./run.sh --skip-vagrant --test-type la --time-limit 20 --nemesis-type partition`
- Increase number of concurrecy: `/bin/bash ./run.sh --skip-vagrant --test-type la --time-limit 20 --concurrency 20 --nemesis-type partition`
- Other arguments:
  - `-r 10` (rate of operations)
  - generator-related args (See [link](https://github.com/jepsen-io/elle/blob/main/src/elle/list_append.clj#L920))
    - `--key-count` (number of distinct keys at any point)
    - `--min-txn-length` (minimum number of operations per txn)
    - `--max-txn-length` (maximum number of operations per txn)
    - `--max-writes-per-key` (maximum number of operations per key)
  - e.g. `/bin/bash ./run.sh --skip-vagrant --test-type la --time-limit 20 -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop`

### And more

Click [here](doc/intro.md) to see the doc of the test(s).

## Acknowledgement

This project is inspired by [jepsen.rqlite](https://github.com/wildarch/jepsen.rqlite). The setup follows a similar procedure.
