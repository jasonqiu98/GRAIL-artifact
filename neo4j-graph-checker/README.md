# Neo4j Graph Checker
This is a Neo4j graph checker; it includes a JMH Benchmark lib to do the experimental test.
## How to use
1. Use docker compose to start to container of neo4j
    ```bash
    docker compose create 
    docker compose start
    ```
2. Select the benchmark you want to run in `grail.CypherBenchmarkTest`, to comment or uncomment the annotation `@Benchmark`
3. Then run `./mvnw clean compile` to build the project
4. Run the class `grail.CypherBenchmarkTest`, it needs an argument as the name of directory