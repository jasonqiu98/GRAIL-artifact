# Dataset

We check the isolation anomalies in a set of execution histories we collected by running test executions on ArangoDB. These test executions consist of a randomly generated set of transactions with a randomly generated set of operation invocations on some list or register variables. To observe isolation anomalies, we run ArangoDB in the cluster setting with replication and sharding enabled for list histories. We collect register histories in a single-node database since ArangoDB supports WAL, which GRAIL uses to infer version histories, currently only in the single-node setting.

We generated the test executions for collecting histories using the Jepsen database testing framework. To evaluate the performance of GRAIL on various workload characteristics, we collected test executions with varying values for the following configurations:

- Duration of history collection (s): Increasing history collection time results in longer execution histories with a higher number of transactions. We used the default value of 100s for collecting the test executions. To analyze the effect of increasing the amount of history collection, we also collected test executions varying the collection time between 10s to 200s (with a step of 10s each time).
- Rate of transaction generation (#transactions/s): Increasing the rate of transaction generation also increases the amount of concurrency. We used the default transaction generation rate of 80 transactions/s for lists and 10 transactions/s for registers. To analyze the effect of its increase, we also collected test executions varying the rate between 10 to 200 (with a step of 10 each time).
- With/without network partitions: Injecting network partitions in the cluster potentially increases the number of isolation anomalies in the execution. We injected random network faults during test executions using Jepsen’s nemesis fault injector. We periodically inject faults that partition the network for five seconds and then we recover the network.

We used the following sets of test execution histories:

- DS1 Executions on lists with a fixed rate of transaction generation, increasing collection time, without network partitions;
- DS2 Executions on lists with a fixed rate of transaction generation, increasing collection time, with network partitions;
- DS3 Executions on lists with an increasing rate of transaction generation, fixed collection time, without network partitions;
- DS4 Executions on lists with an increasing rate of transaction generation, fixed collection time, with network partitions;
- DS5 Executions on lists with a fixed rate of transaction generation, fixed collection time, with network partitions;
- DS6 Executions on registers with a fixed rate of transaction generation, increasing collection time, without network partitions.

We have included the characteristics of the datasets [DS1-DS6] in the following table.

|     | avg. #vertices | avg. #RW | avg. #WW | avg. #WR | avg. density |
|-----|----------------|----------|----------|----------|--------------|
| DS1 | 4399           | 5414     | 3920     | 4356     | 3.11         |
| DS2 | 1643           | 2014     | 1446     | 1616     | 3.09         |
| DS3 | 4557           | 5469     | 3456     | 4089     | 2.99         |
| DS4 | 1624           | 1909     | 1121     | 1374     | 2.86         |
| DS5 | 477            | 586      | 378      | 447      | 2.96         |
| DS6 | 1038           | 1069     | 990      | 1067     | 3.00         |

The table contains the following characteristics.

- Average number of vertices (avg. #vertices), calculated per 20 execution histories of each of [DS1-DS6].
- Average number of each type of edges (avg. #RW | avg. #WW | avg. #WR), calculated per 20 execution histories of each of [DS1-DS6].
- Average density (avg. density), which is the ratio of the number of edges to the number of vertices (called *edge-vertex ratio*), calculated per 20 execution histories of each of [DS1-DS6]. A higher edge-vertex ratio indicates a higher average number of edges connected to each vertex, which leads to a more connected graph.

In the table above, DS2 is the fault-injected version of DS1. Since network partitions interrupt the progress of history generation, DS2 has a fewer number of vertices and edges on average than DS1. This is also true of the pair DS4 and DS3, with DS4 the fault-injected version. Also, the density of the fault-injected version tends to be lower. Furthermore, the table implies that a dependency graph tends to have the least number of WW edges compared with the other two types. This is because the write event on each value, which is specific to a variable, is unique, while this value may be retrieved by multiple read events.

