# ArangoDB Histories with Replication and Sharding

The following configurations apply universally for the following five sets of histories:

- number of threads to generate histories: 10
- key count at the same time: 5
- min txn length: 4
- max txn length: 8
- max writes per key: 8

## `collection-time`

20 histories with fixed rate, increasing collection time, without nemesis

- time-limit: increasing from 10s to 200s with a step of 10s
- rate: 80
- nemesis: none
- replication factor: 3
- sharding factor: 2

## `collection-time-nemesis`

20 histories with fixed rate, increasing collection time, with nemesis

- time-limit: increasing from 10s to 200s with a step of 10s
- rate: 80
- nemesis: random partition into halves
- replication factor: 3
- sharding factor: 2

## `rate`

20 histories with increasing rate, fixed collection time, without nemesis

- time-limit: 100s
- rate: increasing from 10 to 200 with a step of 10
- nemesis: none
- replication factor: 3
- sharding factor: 2

## `rate-nemesis`

20 histories with increasing rate, fixed collection time, with nemesis

- time-limit: 100s
- rate: increasing from 10 to 200 with a step of 10
- nemesis: random partition into halves
- replication factor: 3
- sharding factor: 2

## `histories-30s`

20 histories with fixed rate, fixed collection time, with nemesis, to demonstrate the common usage

- time-limit: 30s
- rate: 80
- nemesis: random partition into halves
- replication factor: 5
- sharding factor: 3

## `rw-register`

`rw-register`: a set of histories (with the accompanying WAL logs) with increasing collection time from 10s to 200s.

- key count at the same time: 5
- min txn length: 4
- max txn length: 8
- max writes per key: 8

- rate: 10
- number of threads to generate histories: 20
- Key-count 5
- Max-writes-per-key 3

`rw-register-test` is not a benchmark, but some prepared test cases to ensure the correctness of the checker.
