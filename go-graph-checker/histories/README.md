# Benchmarks used for Query-Based Checker

1. `list-append`: a set of histories of the Elle's list-append format with increasing collection time from 10s to 300s. Usually, only the histories from 10s to 200s are used in the tests. The `list-append.edn` is not part of the benchmark but a prepared test case to ensure correctness.
   - additional specs
     -  r 10
     - concurrency 20
     - Key-count 5
     - Min-txn-length 4
     - Max-txn-length 8
     - Max-writes-per-key 3

2. `list-append-scalability-rate`: a set of histories of the Elle's list-append format with increasing opreation generation rate from 10 to 300. The files are renamed by dividing the rate by 10.
   - additional specs
     - time-limit 100
     - concurrency 20
     - Key-count 5
     - Min-txn-length 4
     - Max-txn-length 8
     - Max-writes-per-key 3


3. `rw-register`: a set of histories (with the accompanying WAL logs) with increasing collection time from 10s to 300s. Usually, only the histories from 10s to 200s are used in the tests. The `rw-regsiter.edn` (with `rw-register.log`) is not part of the benchmark but a prepared test case to ensure correctness.
   - additional specs
     - r 10
     - concurrency 20
     - Key-count 5
     - Min-txn-length 4
     - Max-txn-length 8
     - Max-writes-per-key 3


4. `rw-register-scalability`

   - `1.edn` to `30.edn`: histories with collection time increasing from 10s to 300s, without any nemesis (same to previous settings).
     - additional specs
       - r 10
       - concurrency 20
       - Key-count 3
       - Min-txn-length 2
       - Max-txn-length 8
       - Max-writes-per-key 32

   - `31.edn` to `60.edn`: histories with collection time increasing from 10s to 300s, with nemesis to partition the network into two halves.
     - additional specs
       - r 10
       - concurrency 20
       - Key-count 3
       - Min-txn-length 2
       - Max-txn-length 8
       - Max-writes-per-key 32

   - `61.edn` to `89.edn`: histories with max transaction length increasing from 2 to 30, with collection time fixed to 100s, without any nemesis.
     - additional specs
       - r 10
       - concurrency 20
       - Key-count 3
       - Min-txn-length 2
       - **time-limit 100**
       - Max-writes-per-key 32

   - `90.edn` to `118.edn`: histories with number of concurrent workers at the same increasing from 2 to 30, with collection time fixed to 100s, without any nemesis, with max transction length fixed to 8.
     - additional specs
       - r 10
       - **time-limit 100*
       - Key-count 3
       - Min-txn-length 2
       - Max-txn-length 8
       - Max-writes-per-key 32

`rw-register-test` is not a benchmark, but some prepared test cases to ensure the correctness of the checker.
