# go-elle

This package is directly downloaded from the project [pingcap/tipocket](https://github.com/pingcap/tipocket/tree/6f8e6014b74d712a4d0c60e4a56c2456285d07b6/pkg/elle), which is a Go implementation of [jepsen-io/elle](https://github.com/jepsen-io/elle).

## I. Elle: List Append Case

- [`Check`](./list_append/list_append.go#L579) detects all the anti-patterns of a history (a slice of ops).
  1. [`preProcessHistory`](./list_append/utils.go#L438) does some necessary pre-processing.
     - [`FilterOutNemesisHistory`](./core/util.go#L15) only keeps non-nemesis ops.
     - [`AttachIndexIfNoExists`](./core/history.go#L301) attaches new indices if the original ones are not recorded or lost for some reason.
  2. [`g1aCases`](./list_append/list_append.go#L352) finds all the aborted reads from a history.
     - [`FailedWrites`](./txn/helper_fn.go#L79) collects all the failed appends in a map with the structure `{k: {v: &op}}`, where `k` and `v` are key and value of each aborted append micro-op, and `&op` is the address of the op that contains the micro-op. This can be regarded as a *version map* from each new value appended to the list (but failed) to the pointer of its corresponding op.
     - [`FilterOkHistory`](./core/util.go#L27) collects all the successful ops in a history.
     - Then this case will start an iteration and find all the read micro-ops that violate the "aborted reads".
       - For each (list) value of the micro-op, all the list elements will be checked. Any element `v` is supposed to have been appended properly to the list, so the version map should not contain any entry with `v` as its key; otherwise, a violation of "aborted read" can be found.
  3. [`g1bCases`](./list_append/list_append.go#L385) finds all the intermediate reads from a history.
     - [`IntermediateWrites`](./txn/helper_fn.go#L53) collects all the adjacent appends in a map with the structure `{k: {v1: &op2}}`, where `k` is the key for both appends, `v1` are the value of the first append, and `&op2` is the address of the following append (that also overwrites the old value). This can be regarded as a *version map* from each new value appended to the list to the pointer to the next append (if exists).
     - [`FilterOkHistory`](./core/util.go#L27) collects all the successful ops in a history.
     - Then this case will start an iteration and find all the read micro-ops that violate the "intermediate reads".
       - For each (list) value of the micro-op, only the latest element will be checked. This latest element `v1` should not be followed by any later appends, so the version map should not contain any entry with `v1` as its key; otherwise, a violation of "intermediate reads" can be found.
  4. [`internalCases`](./list_append/list_append.go#L486) checks whether other internal factors invalidate the history.
     - [`FilterOkHistory`](./core/util.go#L27) collects all the successful ops in a history.
     - [`opInternalCase`](./list_append/list_append.go#L433)
       - This function will first create a *data map* (`{k: [v0 v1 v2 ...]}`) collecting the appended values as a list under each key. Specially, `v0` is a prefix for initialization.
       - Then every time a read micro-op happens, check whether the current record is consistent with the one in the data map. This helps check violations against "read after write".
  5. [`dirtyUpdateCases`](./list_append/list_append.go#L516)
     - This function first extracts the trace of values under each key that we are confident to infer.
       - [`sortedValues`](./list_append/utils.go#L353) finds the trace of list values under each key, removes possible duplicates and builds up a map `{k: [[v1] [v1 v2] [v1 v2 v3] ...]}`. It gives a unique trace of list values ordered by length (and the lengths may be non-consecutive). Some ops in the history, as exceptions, only consist of a single append micro-op. The list values in such cases are also included in the results.
         - My comment: this function drops those ops that can't give direct order information.
       - [`appendIndex`](./list_append/utils.go#L165) reduces the list array into one single list for each map value. The new map structure is `{k: [v1 v2 v3 ...]}`. This reduction is done through [`mergeOrders`](./list_append/utils.go#L309), which finds the common subsequence plus the trailing part of the longer one step by step.
         - My comment: this function doesn't actually include indices, which is the difference than the original Clojure version.
     - As the second step, [`writeIndex`](./list_append/utils.go#L112) filters all the append micro-ops and structures them into a version map `{k: {v: &op}}`. The `&op` is the pointer to each writer of that value, i.e., normally it points to the op with type `:ok` or `:fail`, but sometimes a type `:info` implies a non-deterministic output.
     - After that, the function will check the type of the pair (the current op, the writer), and find possible anomalies caused by dirty updates.
  6. [`duplicates`](./list_append/utils.go#L82) checks duplicates in the ops of info or ok type.
  7. [`incompatiableOrders`](./list_append/utils.go#L418) checks incompatiable orders in the ops of info or ok type (after the values are sorted by `sortedValue`).
  8. [`AdditionalGraphs`](./txn/cycle.go#L221)
  9. [`Cycles`](./txn/txn.go#L42)
     - [`wwGraph`](./list_append/list_append.go#L104), [`wrGraph`](./list_append/list_append.go#L190) and [`rwGraph`](./list_append/list_append.go#L302) rely on the following three index results.
       - `appendIndex`: key to trace
       - `writeIndex`: key to (the only) writer
       - `readIndex`: key to (one or more) readers
     - [`Combine`](./core/depend.go#L74)
     - [`Check`](./core/core.go#L328)
       - Trajan's Algorithm to find strongly connected components in linear time
       - [`StronglyConnectedComponents`](./core/graph.go#L344)
       - [`explainSCC`](./core/depend.go#L287)
         - [`FindCycle`](./core/graph.go#L552)
         - [`NewCircle`](./core/core.go#L119)
