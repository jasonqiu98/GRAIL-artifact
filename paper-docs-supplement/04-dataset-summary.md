# Dataset Summary

We used Jepsen framework to collect six datasets in total [DS1-DS6], each involving 20 randomly generated histories, to form our workload to evaluate the performance of GRAIL. The first five datasets [DS1-DS5] were generated on lists while the last one [DS6] was on registers. The datasets differ in the progress of two main factors, collection time and the rate of transaction generation. Also, each dataset includes progress of only one main factor: [DS1/DS2] increases collection time, while [DS3/DS4] increases rate of transaction generation. In addition, DS1 and DS2 follow the same configuration except that DS2 are with network partitions, while DS1 are not. This is also true of the pair DS3 and DS4. DS5 is a special dataset to observe the effectiveness of GRAIL, which does not vary any main factor but have network partitions.

In particular, we used the following sets of test execution histories:

- DS1 Executions on lists with a fixed rate of transaction generation (80 txns/s), increasing collection time (10s to 200s with a step of 10s), without network partitions;
- DS2 Executions on lists with a fixed rate of transaction generation (80 txns/s), increasing collection time (10s to 200s with a step of 10s), with network partitions;
- DS3 Executions on lists with an increasing rate of transaction generation (10 txns/s to 200 txns/s with a step of 10 txns/s), fixed collection time (100s), without network partitions;
- DS4 Executions on lists with an increasing rate of transaction generation (10 txns/s to 200 txns/s with a step of 10 txns/s), fixed collection time (100s), with network partitions;
- DS5 Executions on lists with a fixed rate of transaction generation (80 txns/s), fixed collection time (100s), with network partitions;
- DS6 Executions on registers with a fixed rate of transaction generation (10 txns/s), increasing collection time (10s to 200s with a step of 10s), without network partitions. 
