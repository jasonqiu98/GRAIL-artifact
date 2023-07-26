#!/bin/bash

## reg-collection-time

for i in {1..20}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit `expr 10 \* $i` -r 80 --concurrency 10 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 8 --nemesis-type noop
done

## reg-rate

for i in {1..20}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 100 -r `expr 10 \* $i` --concurrency 10 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 8 --nemesis-type noop
done

## reg-session

for i in {1..20}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 100 -r 80 --concurrency `expr 10 \* $i` --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 8 --nemesis-type noop
done

## reg-max-write

for i in {1..20}
do
  /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit 100 -r 80 --concurrency 10 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key ${i} --nemesis-type noop
done