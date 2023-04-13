package listappend

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func printLine() {
	fmt.Println("-----------------------------------")
}

func constructArangoGraph(fileName string, t *testing.T) (driver.Database, []int, DBConsts, core.History) {
	dbConsts := DBConsts{
		"starter",    // Host
		8529,         // Port
		"checker_db", // DB
		"txn_g",      // TxnGraph
		"evt_g",      // EvtGraph
		"txn",        // TxnNode
		"a_evt",      // AppendEvtNode
		"r_evt",      // ReadEvtNode
		"dep",        // TxnDepEdge
		"evt_dep",    // EvtDepEdge
	}
	ednFileName := fmt.Sprintf("../histories/list-append/%s.edn", fileName)
	prompt := fmt.Sprintf("Checking %s...", ednFileName)
	fmt.Println(prompt)
	content, err := os.ReadFile(ednFileName)
	if err != nil {
		log.Fatalf("Cannot read edn file %s", ednFileName)
		t.Fail()
	}
	history, err := core.ParseHistory(string(content))
	if err != nil {
		log.Fatalf("Cannot parse edn file %s", ednFileName)
		t.Fail()
	}
	t1 := time.Now()
	db, txnIds := ConstructGraph(txn.Opts{}, history, dbConsts)
	t2 := time.Now()
	constructTime := t2.Sub(t1).Nanoseconds() / 1e6
	fmt.Printf("constructing graph: %d ms\n", constructTime)

	return db, txnIds, dbConsts, history
}

/*
checking serializability:
  - v1: initial ver.
  - v2: break with Golang
  - Pregel: use Pregel SCC algorithm
*/

func TestProfilingSER(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)
		avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckSERV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckSERV2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckSERV3, false)
		avgCheckTimePregel := Profile(db, dbConsts, nil, CheckSERPregel, false)

		fmt.Printf("checking serializability (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - pregel: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimePregel)
		printLine()
	}
}

func TestProfilingSI(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)

		avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckSIV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckSIV2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckSIV3, false)

		fmt.Printf("checking snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3)
		printLine()
	}
}

func TestProfilingPSI(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)

		avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckPSIV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckPSIV2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckPSIV3, false)

		fmt.Printf("checking parallel snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3)
		printLine()
	}
}

// go test -v -timeout 30s -run ^TestListAppendSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendSER(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("170", t)
	valid, cycle := CheckSERV3(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Serializable!")
		PlotCycle(history, cycle, "../images", "la-ser", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

func TestListAppendSERPregel(t *testing.T) {
	db, _, dbConsts, _ := constructArangoGraph("list-append", t)
	CheckSERPregel(db, dbConsts, nil, true)
}

// go test -v -timeout 30s -run ^TestListAppendSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("list-append", t)
	valid, cycle := CheckSIV1(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "la-si", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

// go test -v -timeout 30s -run ^TestListAppendPSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendPSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("list-append", t)
	valid, cycle := CheckPSIV2(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Parallel Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "la-psi", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

func TestExample(t *testing.T) {
	dbConsts := DBConsts{
		"starter",    // Host
		8529,         // Port
		"checker_db", // DB
		"txn_g",      // TxnGraph
		"evt_g",      // EvtGraph
		"txn",        // TxnNode
		"a_evt",      // AppendEvtNode
		"r_evt",      // ReadEvtNode
		"dep",        // TxnDepEdge
		"evt_dep",    // EvtDepEdge
	}
	// borrowed from ../../go-elle/list_append/viz_test.go
	t1 := mustParseOp(`{:index 2, :type :invoke, :value [[:append x 1] [:r y [1]] [:r z [1 2]]]} `)
	t1p := mustParseOp(`{:index 3, :type :ok, :value [[:append x 1] [:r y [1]] [:r z [1 2]]]} `)
	t2 := mustParseOp(`{:index 4, :type :invoke, :value [[:append z 1]]}`)
	t2p := mustParseOp(`{:index 5, :type :ok, :value [[:append z 1]]}`)
	t3 := mustParseOp(`{:index 0, :type :invoke, :value [[:r x [1 2]] [:r z [1]]]}`)
	t3p := mustParseOp(`{:index 1, :type :ok, :value [[:r x [1 2]] [:r z [1]]]}`)
	t4 := mustParseOp(`{:index 6, :type :invoke, :value [[:append z 2] [:append y 1]]}`)
	t4p := mustParseOp(`{:index 7, :type :ok, :value [[:append z 2] [:append y 1]]}`)
	t5 := mustParseOp(`{:index 8, :type :invoke, :value [[:r z nil] [:append x 2]]}`)
	t5p := mustParseOp(`{:index 9, :type :ok, :value [[:r z nil] [:append x 2]]}`)
	h := []core.Op{t3, t3p, t1, t1p, t2, t2p, t4, t4p, t5, t5p}

	db, txnIds := ConstructGraph(txn.Opts{}, h, dbConsts)
	valid, cycle := CheckSERV3(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Serializable!")
		PlotCycle(h, cycle, "../images", "la-ser", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

func mustParseOp(opString string) core.Op {
	op, err := core.ParseOp(opString)
	if err != nil {
		log.Fatalf("expect no error, got %v", err)
	}
	return op
}
