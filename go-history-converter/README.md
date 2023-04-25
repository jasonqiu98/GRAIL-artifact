# Go History Converter

This module borrows the specifications of <https://github.com/anonymous-hippo/PolySI>

This module aims to convert Elle history to the text form of the following. The converted file will be stored under the folder `out` with the same file name and suffix converted from `.edn` to `.txt`. Run functions in `converter_test.go` to get the results.

```
# r means read, w means write
<r/w>(<key>,<value>,<session_id>,<transaction_id>)
```

For example, an Elle transaction in the list-append case

```
{:type :ok, :f :txn, :value [[:append 4 1] [:r 2 []] [:r 4 [1]] [:append 4 2] [:r 2 []] [:r 3 []] [:r 3 []] [:r 4 [1 2]]], :time 18448457710, :process 15, :index 1}
```

will be converted to

```

<r/w>(<key>,<value>,<session_id>,<transaction_id>)
w(4,1,15,1)
r(2,0,15,1)
r(4,1,15,1)
w(4,2,15,1)
r(2,0,15,1)
r(3,0,15,1)
r(3,0,15,1)
r(4,2,15,1)
```

where `:process` goes to the entry of `session_id` and `:index` goes to the entry of `transaction_id`.

