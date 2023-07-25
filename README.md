# GRAIL: Checking Isolation Violations with Graph Queries

Artifact of GRAph-based Isolation Level checking (GRAIL): graph query-based isolation checkers

Note: We present the experimental results in HTML files listed in [`GRAIL-experiments`](./GRAIL-experiments/)

- ArangoDB-Cycle, ArangoDB-SP & ArangoDB-Pregel: in this repo, under the path [`go-graph-checker`](./go-graph-checker/)
- Neo4j-APOC & Neo4j-GDS-SCC: in this repo, under the path [`neo4j-graph-checker`](./neo4j-graph-checker/)

## Collect histories

We use [`jepsen.arangodb`](./jepsen.arangodb) to collect list histories, and [`jepsen.arangodb.single`](./jepsen.arangodb.single) to collect register histories. See the two repositories for details.

## Set up the checker

See details in [`setup.md`](./docs/setup.md). The usage is provided below.

1. Make sure Docker is on. Run `make start`.
2. Open a new terminal window. Run `docker exec -it arango-test sh`.
3. Run the following to get a quick result.

```shell
cd ~/project
go mod tidy
# use -v flag to see the output
go test -v -timeout 30s -run ^TestListAppendSER$ github.com/grail/anti-pattern-graph-checker-single/go-graph-checker/list_append
```

## Experiments

### I. Effectiveness and Scalability of GRAIL

1. Follow the instructions of [`setup.md`](./docs/setup.md) to get the results of ArangoDB-Cycle, ArangoDB-SP, and ArangoDB-Pregel
   - First rename the value of variable `endFileName` in the function `constructArangoGraph`, in [`list_append_test.go`](./go-graph-checker/list_append/list_append_test.go) or [`rw_register_test.go`](./go-graph-checker/rw_register/rw_register_test.go), respectively. The value of this variable should contain the pattern of the filenames of the histories in one directory ready for tests.
   - Run `TestCorrectness` to get effectiveness results, in `list_append_test.go` or `rw_register_test.go`, respectively.
   - Run `TestProfilingScalability` to get scalability results, in `list_append_test.go` or `rw_register_test.go`, respectively.

2. Get the results from the path [`neo4j-graph-checker`](./neo4j-graph-checker/) by following the steps below.
   - Run `docker compose up` to start the Neo4j Docker container
   - Select the benchmarking functions you want to run in `grail.CypherBenchmarkTest` by commenting or uncommenting the annotation `@Benchmark`
     - For Neo4j-APOC, uncomment the following benchmarks and comment others: [`SerTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L39), [`SITest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L46), [`PSITest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L78), [`PL2Test()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L96), [`PL1Test()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L103)
     - For Neo4j-GDS-SCC, uncomment the following benchmarks and comment others: [`Q1SerProjTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L110), [`Q2SerSCCTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L118), [`Q3_SISCCTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L125), [`Q4PSISCCTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L160), [`Q5PL2ProjTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L180), [`Q6PL2SCCTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L188), [`Q7PL1ProjTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#L195), [`Q8PL1SCCTest()`](./neo4j-graph-checker/src/test/java/grail/CypherBenchmarkTest.java#LL203)
   - `mvn clean install compile` to install dependencies
   - Run the test class `grail.CypherBenchmarkTest` and pass in the directory of the histories

### II. Comparison with Elle

To run `elle-cli` benchmarking on a history, follow the steps below.

1. Clone project

```bash
git clone https://github.com/ligurio/elle-cli
cd elle-cli
```

2. Make the following changes on the repo.

   - Insert the following code above line 236 of [`cli.clj`](https://github.com/ligurio/elle-cli/blob/a36790c1973360792ce455282894fadaeb58796a/src/elle_cli/cli.clj#L236).

   ```clojure
   t             (criterium/quick-bench (check-history model-name history options))
   ```

   - Import the namespace in line 4-23 of [`cli.clj`](https://github.com/ligurio/elle-cli/blob/a36790c1973360792ce455282894fadaeb58796a/src/elle_cli/cli.clj#L4)

   ```clojure
   [criterium.core :as criterium]
   ```

   - Add the dependency in [`project.clj`](https://github.com/ligurio/elle-cli/blob/a36790c1973360792ce455282894fadaeb58796a/project.clj).

   ```clojure
   [criterium "0.4.6"] ; for benchmarking https://github.com/hugoduncan/criterium
   ```

3. Start the checker on one history

```bash
$ lein deps
$ lein uberjar
Compiling elle_cli.cli
Compiling elle_cli.cli
Created /home/grail/thesis-workspace/GRAIL-artifact/elle-cli/target/elle-cli-0.1.6.jar
Created /home/grail/thesis-workspace/GRAIL-artifact/elle-cli/target/elle-cli-0.1.6-standalone.jar
$ java -jar target/elle-cli-0.1.6-standalone.jar --model list-append --consistency-models serializable ../go-graph-checker/histories/collection-time/10.edn
Evaluation count : 30 in 6 samples of 5 calls.
             Execution time mean : 24.489192 ms
    Execution time std-deviation : 1.962274 ms
   Execution time lower quantile : 21.845343 ms ( 2.5%)
   Execution time upper quantile : 26.362684 ms (97.5%)
                   Overhead used : 2.130864 ns
../go-graph-checker/histories/collection-time/10.edn     false
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
sudo apt update
sudo apt install g++ openjdk-11-jdk cmake libgmp-dev zlib1g-dev
git clone --recurse-submodules https://github.com/amnore/PolySI
cd PolySI
./gradlew jar
```

4. Run PolySI on each history converted by `go-history-converter`. We have provided the converted text histories from our datasets.

```bash
java -jar build/libs/PolySI-1.0.0-SNAPSHOT.jar audit --type=text ../go-history-converter/collection-time/10.txt &> polysi.log
```

An example log is shown below.

```
Sessions count: 11
Transactions count: 435
Events count: 2827
Known edges: 1008
Constraints count: 1143
Total edges in constraints: 5832
Pruning round 1
Before: 1008 edges
After: 787 edges
reachability matrix sparsity: 0.53
solved 885 constraints
Pruning round 2
Before: 2785 edges
After: 1770 edges
reachability matrix sparsity: 0.51
solved 65 constraints
Pruning round 3
Before: 2917 edges
After: 1850 edges
reachability matrix sparsity: 0.51
solved 5 constraints
Pruned 3 rounds, solved 955 constraints
After prune: graphA: 1555, graphB: 894
After Prune:
Constraints count: 188
Total edges in constraints: 593
Before: 2919 edges
After: 1851 edges
Graph A edges count: 1965
Graph B edges count: 1077
Graph A union C edges count: 3204
ENTIRE_EXPERIMENT: 292ms
ONESHOT_CONS: 40ms
SI_VERIFY_INT: 11ms
SI_GEN_PREC_GRAPH: 16ms
SI_GEN_CONSTRAINTS: 10ms
SI_PRUNE: 87ms
SI_PRUNE_POST_GRAPH_A_B: 40ms
SI_PRUNE_POST_GRAPH_C: 3ms
SI_PRUNE_POST_REACHABILITY: 35ms
SI_PRUNE_POST_CHECK: 5ms
ONESHOT_SOLVE: 139ms
SI_SOLVER_GEN: 67ms
SI_SOLVER_GEN_GRAPH_A_B: 7ms
SI_SOLVER_GEN_REACHABILITY: 10ms
SI_SOLVER_GEN_GRAPH_A_UNION_C: 27ms
SI_SOLVER_GEN_MONO_GRAPH: 19ms
SI_SOLVER_SOLVE: 3ms
Max memory: 29.4MB
[[[[ ACCEPT ]]]]
```
