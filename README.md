# Anti-Pattern Graph Checker (Single Instance)

A graph-based checker to detect anti-patterns of a database

## I. Quickstart the graph checker

### 0. Specifications

- OS: Linux Mint 21
- Docker Engine - Community 20.10.22
- Docker Compose version v2.14.1
- Within `arango-test` container
  - ArangoDB v3.9.2
  - Go 1.19.4
  - GCC (latest in use)

### 1. Set up the Docker containers for history processing

1. Make sure the Docker daemon is running on your machine.
2. Start the orchestration by `docker compose up` (or in the backend: `docker compose up -d`), until you see some messages like the following.

```
arango-starter  | 1677539493.202554 [1] INFO [cf3f4] {general} ArangoDB (version 3.9.2 [linux]) is ready for business. Have fun!
arango-test     | 2023-02-27T23:11:33Z [1] INFO [cf3f4] {general} ArangoDB (version 3.9.2 [linux]) is ready for business. Have fun!
```

3. Open a new terminal and enter the `arango-test` container by `docker exec -it arango-test sh`.

### 2. Run the project at "~/project" within the `arango-test` Docker container

Following the above instructions to set up the environment, you can run this Go project. The project path has been linked to the path `~/project` in the Docker container `arango-test`. ArangoDB, Go and GCC have been pre-installed.

```shell
cd ~/project
go mod tidy
# use -v flag to see the output
go test -v -timeout 30s -run ^TestListAppendSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
go test -v -timeout 30s -run ^TestExample$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
```

### 3. Profiling

Use the following command to profile different approaches.

```shell
go test -v -timeout 300s -run ^TestProfilingSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
go test -v -timeout 360s -run ^TestProfilingSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
go test -v -timeout 360s -run ^TestProfilingPSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
```

### 3. Stop and remove the containers

Open a new terminal on you local machine, and run the following commands.

- Stop
  - `docker compose stop` to stop all the services
- Remove
  - `docker compose down -v` or `docker compose down --volumes` to remove all the containers, networks and volumes generated by the `docker-compose.yml` file
  - `docker rmi arangodb:3.9.2` and `docker rmi anti-pattern-graph-checker-single-test:latest` to remove the relevant Docker images.
- Clean and remove the persisting volumes
  - Run `sudo rm -rf nodes logs` to delete relevant records/logs (if you don't need them for analysis any more)

### 4. Q&A

1. How do I resolve the error message like `dial tcp: lookup starter: no such host`?

- This may happen when you run the program on your local environment, which is not well connected to our Docker cluster. To resolve this problem, you can either
  - open a new terminal and enter the `arango-test` container by `docker exec -it arango-test sh` (recommended), or
  - access `http://localhost:8000` from the local environment

2. How do I connect to the Web interface of ArangoDB (on local machine)?

- Access `http://localhost:8000` in the browser. Also remember to choose the correct database you are trying to access.

3. How do I connect to the database using `arangosh` (in `arango-test` container)?

- Connect to the default database `_system` via the ArangoDB starter, by `arangosh --server.endpoint tcp://starter:8529 --server.username root --server.password ""`, and then submit any JavaScript commands.

4. How do I access the logs (on local machine)?

- Run `sudo chmod 777 logs/*/*.log` on the root of the project folder before accessing the log files on your local machine.

5. How do I read the logs?

- The logs are quite verbose. This is because (1) the logs have been set to TRACE level, meaning the log of all levels will be written to the log files, and (2) both user queries and system executions will be recorded in the logs. If you want to see a shorter log file, you can adjust the `--log.level` options in the `docker-compose.yml` file to a level that is satisfactory to you.
- Read the logs and understand the execution of each query of your interest. Try searching code `[11160]` (meaning query begins) and code `[f5cee]` (meaning query ends) in `queries.log`.

6. How do I remove the cached test results in Golang?

- `go clean -testcache`: expires all test results

## II. Quickstart by `Makefile` [Optional]

If you have `gcc` installed on your machine, you can use the following commands to run this project.

- `make start`: start the orchestration
- `make stop`: stop the orchestration
- `make remove`: remove the orchestration and Docker images
- `make clean`: remove the local persisting volumes of the database

It is still recommended to run the raw commands by yourself, but `Makefile` can be a help.

## III. Run Elle tests (with list-append cases)

Run the following command (locally or in the `arango-test` container).

```shell
go test -timeout 30s -run "^(TestWWGraph|TestWRGraph|TestGraph|TestG1aCases|TestG1bCases|TestInternalCases|TestChecker|TestRepeatableRead|TestGNonadjacent|TestCheck|TestHugeScc|TestPlotAnalysis)$" github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/list_append
```
