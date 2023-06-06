# ArangoDB Histories

## List Histories with Replication and Sharding

The following configurations apply universally for all list histories.

- number of sessions (number of threads to generate historiess): 10
- key count at the same time: 5
- min txn length: 4
- max txn length: 8
- max writes per key: 8

### [List1] `list-collection-time`

20 histories with fixed rate, increasing collection time, without nemesis

- time-limit: increasing from 10s to 200s with a step of 10s
- rate: 80
- nemesis: none
- replication factor: 3
- sharding factor: 2

### [List2] `list-collection-time-nemesis`

20 histories with fixed rate, increasing collection time, with nemesis

- time-limit: increasing from 10s to 200s with a step of 10s
- rate: 80
- nemesis: random partition into halves
- replication factor: 3
- sharding factor: 2

### [List3] `list-rate`

20 histories with increasing rate, fixed collection time, without nemesis

- time-limit: 100s
- rate: increasing from 10 to 200 with a step of 10
- nemesis: none
- replication factor: 3
- sharding factor: 2

### [List4] `list-rate-nemesis`

20 histories with increasing rate, fixed collection time, with nemesis

- time-limit: 100s
- rate: increasing from 10 to 200 with a step of 10
- nemesis: random partition into halves
- replication factor: 3
- sharding factor: 2

### [List5] `list-histories-30s`

20 histories with fixed rate, fixed collection time, with nemesis, to demonstrate the common usage

- time-limit: 30s
- rate: 80
- nemesis: random partition into halves
- replication factor: 5
- sharding factor: 3

## Register Histories (without Sharding or Replication)

The following configurations apply universally for all list histories.

- key count at the same time: 5
- min txn length: 4
- max txn length: 8

### [Reg1] `reg-collection-time`

a set of histories (with the accompanying WAL logs) with increasing collection time from 10s to 200s

- time-limit: increasing from 10s to 200s with a step of 10s
- rate: 20
- number of sessions (number of threads to generate historiess): 10
- max writes per key: 8

### [Reg2] `reg-rate`

a set of histories (with the accompanying WAL logs) with increasing rate from 10 to 200

- time-limit: 30s
- rate: increasing from 10 to 200 with a step of 10
- number of sessions (number of threads to generate historiess): 10
- max writes per key: 8

### [Reg3] `reg-session`

a set of histories (with the accompanying WAL logs) with increasing number of sessions from 10 to 200

- time-limit: 30s
- rate: 20
- number of sessions (number of threads to generate historiess): increasing from 10 to 200 with a step of 10
- max writes per key: 8

### [Reg4] `rw-register`

a set of histories (with the accompanying WAL logs) with increasing max writes per key from 1 to 20

- time-limit: 30s
- rate: 20
- number of sessions (number of threads to generate historiess): 10
- max writes per key: increasing from 10 to 200 with a step of 1

## Not part of benchmarks

- `anti-patterns` shows two anti-pattern examples used in the thesis
- `rw-register-test` is not a benchmark, but some prepared test cases for the effectiveness of the checker.
