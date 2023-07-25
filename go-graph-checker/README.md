# Query-Based Checker

1. Collect history
   1. list-append case: from the GitHub repository [`jepsen.arangodb`](https://github.com/grail/jepsen.arangodb).
      - For example
        - ```bash
          #!/bin/bash
          for t in {10..200..10}
          do
            /bin/bash ./run.sh --skip-vagrant --test-type la --time-limit ${t} -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop
          done
          ```
   2. rw-register case: from the GitHub repository [`jepsen.arangodb.single`](https://github.com/grail/jepsen.arangodb.single).
      - For example
        - ```bash
          #!/bin/bash
          for t in {10..200..10}
          do
            /bin/bash ./run.sh --skip-vagrant --test-type rw --time-limit ${t} -r 10 --concurrency 20 --key-count 5 --min-txn-length 4 --max-txn-length 8 --max-writes-per-key 3 --nemesis-type noop
          done
          ```
2. Follow the project instructions to start the docker. Remember to run all the programs **within** the docker container by running the command `docker exec -it arango-test sh`.
3. Run the following commands to install any dependencies.

```bash
cd ~/project
go mod tidy
```

4. Run any test (including functionality, profiling or benchmark) with command line or VS Code Golang plugins. For example, the following command tests the correctness of the checker, following the TDD principles.

```bash
go test -v -timeout 120s -run ^TestChecker$ github.com/grail/anti-pattern-graph-checker-single/go-graph-checker/list_append
```
