# Anti-Patterns

to record two typical anti-patterns detected by the query-based checker

1. `write-skew`
    - replication: 1; sharding: 1 (i.e., no replication and sharding)
    - command: `lein run test --nodes-file ./vagrant/nodes-vagrant.txt --password --ssh-private-key ./vagrant/vagrant_ssh_key --username vagrant --test-type la --time-limit 120 -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop`
    - Output by Query-Based Checker (by `TestCheckExample`)
        ```
        2023/05/21 13:24:52 Checking ../histories/anti-patterns/write-skew/history.edn...
        2023/05/21 13:24:52 db exists already and will be dropped first...
        constructing graph: 6763 ms
        2023/05/21 13:24:59 Cycle Detected by SER V3.
        2023/05/21 13:24:59 T154 (rw) T155 (rw) T154
        2023/05/21 13:24:59 Not Serializable!
        2023/05/21 13:24:59 Anti-Patterns for Snapshot Isolation Not Detected.
        2023/05/21 13:24:59 Anti-Patterns for Parallel Snapshot Isolation Not Detected.
        2023/05/21 13:25:00 Anti-Patterns for PL-2 Not Detected.
        2023/05/21 13:25:00 Anti-Patterns for PL-1 Not Detected.
        ```

2. `lost-update`
    - replication: 3; sharding: 2
    - command: `lein run test --nodes-file ./vagrant/nodes-vagrant.txt --password --ssh-private-key ./vagrant/vagrant_ssh_key --username vagrant --test-type la --time-limit 120 -r 80 --concurrency 10 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 8 --nemesis-type noop`
    - Output by Query-Based Checker (by `TestCheckExample`)
        ```
        2023/05/21 13:25:57 Checking ../histories/anti-patterns/lost-update/history.edn...
        2023/05/21 13:25:58 db exists already and will be dropped first...
        constructing graph: 30727 ms
        2023/05/21 13:26:30 Cycle Detected by SER V3.
        2023/05/21 13:26:30 T10168 (rw) T10169 (rw) T10168
        2023/05/21 13:26:30 Not Serializable!
        2023/05/21 13:26:33 Cycle Detected by SI V3.
        2023/05/21 13:26:33 T13379 (wr) T13381 (rw) T13379
        2023/05/21 13:26:33 Not Snapshot Isolation!
        2023/05/21 13:26:35 Cycle Detected by PSI V3.
        2023/05/21 13:26:35 T13379 (wr) T13381 (rw) T13379
        2023/05/21 13:26:35 Not Parallel Snapshot Isolation!
        2023/05/21 13:26:37 Anti-Patterns for PL-2 Not Detected.
        2023/05/21 13:26:38 Anti-Patterns for PL-1 Not Detected.
        ```
