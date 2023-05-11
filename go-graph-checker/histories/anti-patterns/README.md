# Anti-Patterns

to record anti-patterns that violate an isolation level below the claimed Snapshot Isolation ("REPETABLE READ" claimed by ArangoDB)

1. `lost-update-1`
    - replication: 3; sharding: 2
    - command: `lein run test --nodes-file ./vagrant/nodes-vagrant.txt --password --ssh-private-key ./vagrant/vagrant_ssh_key --username vagrant --test-type la --time-limit 30 -r 60 --concurrency 10 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop`
    - Output by Jepsen
        ```
        G-single #0
        Let:
        T1 = {:type :ok, :f :txn, :value [[:r 296 []] [:r 298 [1 2 3]] [:r 299 []] [:append 295 3] [:r 295 [1 2 3]]], :time 24272943316, :process 5, :index 684}
        T2 = {:type :ok, :f :txn, :value [[:r 295 [1 2]] [:r 298 [1 2]] [:append 296 1] [:r 295 [1 2]] [:append 298 3] [:r 298 [1 2 3]] [:r 296 [1]]], :time 24262462607, :process 6, :index 682}

        Then:
        - T1 < T2, because T1 observed the initial (nil) state of 296, which T2 created by appending 1.
        - However, T2 < T1, because T1 observed T2's append of 3 to key 298: a contradiction!
        ```
    - Output by Query-Based Checker (by `TestCheckExample`)
        ```
        2023/05/11 23:01:56 Checking ../histories/anti-patterns/lost-update-1/history.edn...
        2023/05/11 23:01:56 db exists already and will be dropped first...
        constructing graph: 5852 ms
        2023/05/11 23:02:02 Cycle Detected by SER V3.
        2023/05/11 23:02:02 T1017 (rw) T1018 (rw) T1017
        2023/05/11 23:02:02 Not Serializable!
        2023/05/11 23:02:02 Cycle Detected by SI V3.
        2023/05/11 23:02:02 T682 (wr) T684 (rw) T682
        2023/05/11 23:02:02 Not Snapshot Isolation!
        2023/05/11 23:02:02 Cycle Detected by PSI V3.
        2023/05/11 23:02:02 T682 (wr) T684 (rw) T682
        2023/05/11 23:02:02 Not Parallel Snapshot Isolation!
        2023/05/11 23:02:02 Anti-Patterns for PL-2 Not Detected.
        2023/05/11 23:02:02 Anti-Patterns for PL-1 Not Detected.
        ```

2. `lost-update-2`
    - replication: 3; sharding: 2
    - command: `lein run test --nodes-file ./vagrant/nodes-vagrant.txt --password --ssh-private-key ./vagrant/vagrant_ssh_key --username vagrant --test-type la --time-limit 30 -r 60 --concurrency 10 --key-count 8 --min-txn-length 6 --max-txn-length 12 --max-writes-per-key 5 --nemesis-type noop`
    - Output by Jepsen
        ```
        G-single #0
        Let:
        T1 = {:type :ok, :f :txn, :value [[:r 1406 [1 2 3]] [:r 1406 [1 2 3]] [:append 1406 4] [:append 1406 5] [:r 1405 []] [:r 1405 []] [:r 1405 []]], :time 46312979751, :process 4, :index 3393}
        T2 = {:type :ok, :f :txn, :value [[:r 1403 [1 2]] [:append 1404 2] [:r 1404 [2]] [:r 1405 []] [:append 1406 1] [:r 1405 []] [:append 1405 4] [:append 1406 2] [:append 1406 3]], :time 46295588172, :process 2, :index 3391}

        Then:
        - T1 < T2, because T1 observed the initial (nil) state of 1405, which T2 created by appending 4.
        - However, T2 < T1, because T1 appended 4 after T2 appended 3 to 1406: a contradiction!
        ```
    - Output by Query-Based Checker (by `TestCheckExample`)
        ```
        2023/05/11 23:06:28 Checking ../histories/anti-patterns/lost-update-2/history.edn...
        2023/05/11 23:06:28 db exists already and will be dropped first...
        constructing graph: 5855 ms
        2023/05/11 23:06:34 Cycle Detected by SER V3.
        2023/05/11 23:06:34 T112 (wr) T118 (rw) T115 (rw) T112
        2023/05/11 23:06:34 Not Serializable!
        2023/05/11 23:06:34 Cycle Detected by SI V3.
        2023/05/11 23:06:34 T3391 (wr) T3393 (rw) T3391
        2023/05/11 23:06:34 Not Snapshot Isolation!
        2023/05/11 23:06:34 Cycle Detected by PSI V3.
        2023/05/11 23:06:34 T3391 (wr) T3393 (rw) T3391
        2023/05/11 23:06:34 Not Parallel Snapshot Isolation!
        2023/05/11 23:06:34 Anti-Patterns for PL-2 Not Detected.
        2023/05/11 23:06:34 Anti-Patterns for PL-1 Not Detected.
        ```
