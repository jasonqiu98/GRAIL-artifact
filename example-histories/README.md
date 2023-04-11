# README

These histories are collected through the following command from the GitHub repository `jepsen.arangodb`.

## `list-append`

```bash
#!/bin/bash
for i in {1..20}
do
  /bin/bash ./run.sh --skip-vagrant --test-type la --time-limit `expr 10 \* $i` -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop
done
```

## `rw-register`

```shell
for t in {10..200..10}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit $t -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop
done
```
