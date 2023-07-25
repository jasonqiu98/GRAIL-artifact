# The usage of indexing in ArangoDB

ArangoDB automatically uses indexing in its cycle detection queries. For example, when a graph traversal is initiated, the id's of the starting vertices will be used as an index, which is similar to the concept of primary keys of the relational databases. So, in general, we do not have the cases without indexing.