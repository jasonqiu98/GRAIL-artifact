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

## `list-append-scalability`

```bash
#!/bin/bash

for i in {1..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type la --time-limit `expr 10 \* $i` -r 10 --concurrency 20 --key-count 3 --min-txn-length 2 --max-txn-length 8 --max-writes-per-key 32 --nemesis-type noop
done

for i in {1..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type la --time-limit `expr 10 \* $i` -r 10 --concurrency 20 --key-count 3 --min-txn-length 2 --max-txn-length 8 --max-writes-per-key 32 --nemesis-type partition
done

for i in {2..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type la --time-limit 100 -r 10 --concurrency 20 --key-count 3 --min-txn-length 2 --max-txn-length $i --max-writes-per-key 32 --nemesis-type noop
done

for i in {2..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type la --time-limit 100 -r 10 --concurrency `expr 10 \* $i` --key-count 3 --min-txn-length 2 --max-txn-length 8 --max-writes-per-key 32 --nemesis-type noop
done
```

## `rw-register-scalability`

```bash
#!/bin/bash

for i in {1..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit `expr 10 \* $i` -r 10 --concurrency 20 --key-count 3 --min-txn-length 2 --max-txn-length 8 --max-writes-per-key 32 --nemesis-type noop
done

for i in {1..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit `expr 10 \* $i` -r 10 --concurrency 20 --key-count 3 --min-txn-length 2 --max-txn-length 8 --max-writes-per-key 32 --nemesis-type partition
done

for i in {2..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 100 -r 10 --concurrency 20 --key-count 3 --min-txn-length 2 --max-txn-length $i --max-writes-per-key 32 --nemesis-type noop
done

for i in {2..30}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 100 -r 10 --concurrency `expr 10 \* $i` --key-count 3 --min-txn-length 2 --max-txn-length 8 --max-writes-per-key 32 --nemesis-type noop
done
```