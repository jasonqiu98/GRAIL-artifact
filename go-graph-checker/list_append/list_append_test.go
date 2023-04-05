package listappend

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

/*
go test -v -run ^TestProfiling$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
checking serializability:
  - v1: initial ver.
  - v2: break with Golang
  - Pregel: use Pregel SCC algorithm
*/
func TestProfilingSER(t *testing.T) {
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
	fmt.Println("-----------------------------------")
	for d := 10; d <= 200; d += 10 {
		fileName := fmt.Sprintf("../histories/%d.edn", d)
		prompt := fmt.Sprintf("Checking %s...", fileName)
		fmt.Println(prompt)
		content, err := os.ReadFile(fileName)
		if err != nil {
			t.Fail()
		}
		history, err := core.ParseHistory(string(content))
		if err != nil {
			t.Fail()
		}
		t1 := time.Now()
		db, metadata := ConstructGraph(txn.Opts{}, history, dbConsts)
		t2 := time.Now()
		constructTime := t2.Sub(t1).Nanoseconds() / 1e6
		fmt.Printf("constructing graph: %d ms\n", constructTime)

		avgCheckTimeV1 := Profile(db, dbConsts, metadata, CheckSERV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, metadata, CheckSERV2, false)
		avgCheckTimePregel := Profile(db, dbConsts, nil, CheckSERPregel, false)

		fmt.Printf("checking serializability (on avg.):\n - v1: %d ms\n - v2: %d ms\n - pregel: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimePregel)
		fmt.Println("-----------------------------------")
	}
}

func TestProfilingSI(t *testing.T) {
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
	fmt.Println("-----------------------------------")
	for d := 10; d <= 200; d += 10 {
		fileName := fmt.Sprintf("../histories/%d.edn", d)
		prompt := fmt.Sprintf("Checking %s...", fileName)
		fmt.Println(prompt)
		content, err := os.ReadFile(fileName)
		if err != nil {
			t.Fail()
		}
		history, err := core.ParseHistory(string(content))
		if err != nil {
			t.Fail()
		}
		t1 := time.Now()
		db, metadata := ConstructGraph(txn.Opts{}, history, dbConsts)
		t2 := time.Now()
		constructTime := t2.Sub(t1).Nanoseconds() / 1e6
		fmt.Printf("constructing graph: %d ms\n", constructTime)

		avgCheckTimeV1 := Profile(db, dbConsts, metadata, CheckSIV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, metadata, CheckSIV2, false)

		fmt.Printf("checking snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2)
		fmt.Println("-----------------------------------")
	}
}

func TestProfilingPSI(t *testing.T) {
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
	fmt.Println("-----------------------------------")
	for d := 10; d <= 200; d += 10 {
		fileName := fmt.Sprintf("../histories/%d.edn", d)
		prompt := fmt.Sprintf("Checking %s...", fileName)
		fmt.Println(prompt)
		content, err := os.ReadFile(fileName)
		if err != nil {
			t.Fail()
		}
		history, err := core.ParseHistory(string(content))
		if err != nil {
			t.Fail()
		}
		t1 := time.Now()
		db, metadata := ConstructGraph(txn.Opts{}, history, dbConsts)
		t2 := time.Now()
		constructTime := t2.Sub(t1).Nanoseconds() / 1e6
		fmt.Printf("constructing graph: %d ms\n", constructTime)

		avgCheckTimeV1 := Profile(db, dbConsts, metadata, CheckPSIV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, metadata, CheckPSIV2, false)

		fmt.Printf("checking parallel snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2)
		fmt.Println("-----------------------------------")
	}
}

func TestListAppendSER(t *testing.T) {
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
	content, err := os.ReadFile("../histories/120.edn")
	if err != nil {
		t.Fail()
	}
	history, err := core.ParseHistory(string(content))
	if err != nil {
		t.Fail()
	}
	db, metadata := ConstructGraph(txn.Opts{}, history, dbConsts)
	valid := CheckSERV2(db, dbConsts, metadata, true)
	if !valid {
		fmt.Println("Not Serializable!")
	}
	// db, _ := ConstructGraph(txn.Opts{}, history, dbConsts)
	// CheckSERPregel(db, dbConsts, nil, true)
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

	db, metadata := ConstructGraph(txn.Opts{}, h, dbConsts)
	valid := CheckSERV2(db, dbConsts, metadata, true)
	if !valid {
		fmt.Println("Not Serializable!")
	}
}

func mustParseOp(opString string) core.Op {
	op, err := core.ParseOp(opString)
	if err != nil {
		log.Fatalf("expect no error, got %v", err)
	}
	return op
}
