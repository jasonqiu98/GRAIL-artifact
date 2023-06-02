# GRAIL: Checking Isolation Violations with Graph Queries

Artifact of GRAph-based Isolation Level checking (GRAIL): graph query-based isolation checkers

- ArangoDB-Cycle, ArangoDB-SP & ArangoDB-Pregel: in this repo, under the path [`go-graph-checker`](./go-graph-checker/)
- Neo4j-APOC & Neo4j-GDS-SCC: in the repo [`pbt-benchmark`](https://github.com/JINZhao2000/pbt-benchmark)

## Set up the checker

See details in [`setup.md`](./docs/setup.md). The usage is provided below.

1. Make sure Docker is on. Run `make start`.
2. Open a new terminal window. Run `docker exec -it arango-test sh`.
3. Run the following to get a quick result.

```shell
cd ~/project
go mod tidy
# use -v flag to see the output
go test -v -timeout 30s -run ^TestListAppendSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
```

## Experiments

### I. Effectiveness and Scalability of GRAIL

1. Follow the instructions of [`setup.md`](./docs/setup.md) to get the results of ArangoDB-Cycle, ArangoDB-SP, and ArangoDB-Pregel
   - Run `TestCorrectness` to get effectiveness results
   - Run `TestProfilingScalability` to get scalability results

2. Get the results from the repo [`pbt-benchmark`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f) by following the steps below.
   - Run `docker compose up` to start the Neo4j Docker container
   - Select the benchmarking functions you want to run in `cyou.zhaojin.CypherBenchmarkTest` by commenting or uncommenting the annotation `@Benchmark`
     - For Neo4j-APOC, uncomment the following benchmarks and comment others: [`SerTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L43), [`SITest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L50), [`PSITest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L82), [`PL2Test()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L100), [`PL1Test()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L107)
     - For Neo4j-GDS-SCC, uncomment the following benchmarks and comment others: [`Q1SerProjTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L114), [`Q2SerSCCTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L122), [`Q3_SISCCTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L129), [`Q4PSISCCTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L164), [`Q5PL2ProjTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L184), [`Q6PL2SCCTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L192), [`Q7PL1ProjTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#L199), [`Q8PL1SCCTest()`](https://github.com/JINZhao2000/pbt-benchmark/blob/20a236e15e76c91ec5b7cec8b2a3359ae325154f/src/test/java/cyou/zhaojin/CypherBenchmarkTest.java#LL207C17-L207C29)
   - `mvn clean install compile` to install dependencies
   - Run the test class `cyou.zhaoin.CypherBenchmarkTest` and pass in the directory of the histories

### II. Comparison with Elle

#### Usage

1. Clone the repo [elle-cli](https://github.com/ligurio/elle-cli/tree/master).

2. Replace the lines 234-237 of the code file [`cli.clj`](https://github.com/ligurio/elle-cli/blob/a36790c1973360792ce455282894fadaeb58796a/src/elle_cli/cli.clj#L234) with the following code.

```clojure
(let [read-history  (or read-history (read-fn-by-extension filepath))
     history       (h/history (read-history filepath))
     t             (criterium/bench (check-history model-name history options))
     analysis      (check-history model-name history options)
     validness     (:valid? analysis)]
 (println (/ (first t) 1e6) "ms") ; ns to ms
```

3. Pass the directory into the command line with `elle-cli`.

The steps above are presented in the following code.

```bash
$ git clone https://github.com/ligurio/elle-cli
$ cd elle-cli
$ lein deps
$ lein uberjar
Compiling elle_cli.cli
Created /home/sergeyb/sources/ljepsen/elle-cli/target/elle-cli-0.1.6.jar
Created /home/sergeyb/sources/ljepsen/elle-cli/target/elle-cli-0.1.6-standalone.jar
$ java -jar target/elle-cli-0.1.6-standalone.jar --model list-append --consistency-models serializable histories/collection-time/10.edn
259.725797 ms
histories/replication/collection-time/10.edn     false
$ java -jar target/elle-cli-0.1.6-standalone.jar --model list-append --consistency-models snapshot-isolation histories/collection-time/10.edn
264.391106 ms
histories/replication/collection-time/10.edn     true
$ java -jar target/elle-cli-0.1.6-standalone.jar --model list-append --consistency-models serializable histories/collection-time/90.edn
887.769591 ms
histories/replication/collection-time/80.edn     false
$ java -jar target/elle-cli-0.1.6-standalone.jar --model list-append --consistency-models snapshot-isolation histories/collection-time/90.edn
684.537614 ms
histories/replication/collection-time/80.edn     false
```

### III. Comparison with PolySI

1. Make directories within [`go-history-converter`](./go-history-converter) as follows.
   - to convert list histories: `mkdir -p list-append`
   - to convert register histories: `mkdir -p rw-register`

2. Convert the `.edn` histories to plain-text histories.
   - by function `testConverter` in [`convert_test.go`](./go-history-converter/converter_test.go)
     - The function takes three arguments as follows: the file name of the history `path`, a boolean variable `rw` to indicate an RW-register history (`True`) or not (`False`), and a fixed argument for Go testing `t`.
     - Users need to modify the path, linking it to the correct file.
   - by function `TestConvertsAll` in `convert_test.go`
     - The function iterates over all files to convert them.
     - Users need to define the loop, modify the path of the directory, and also toggle the `rw` boolean variable if necessary
  
3. Clone PolySI

```bash
$ sudo apt update
$ sudo apt install g++ openjdk-11-jdk cmake libgmp-dev zlib1g-dev
$ git clone --recurse-submodules https://github.com/amnore/PolySI
$ cd PolySI
$ ./gradlew jar
```

4. Run PolySI on histories in `go-history-converter`. We have provided the converted text histories from our datasets.

```bash
$ java -jar build/libs/PolySI-1.0.0-SNAPSHOT.jar audit --type=text ../go-history-converter/collection-time/10.txt &> polysi.log
```
