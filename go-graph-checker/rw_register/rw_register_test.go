package rwregister

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func printLine() {
	fmt.Println("-----------------------------------")
}

func constructGraph(fileName string, t *testing.T) (driver.Database, map[string]int, DBConsts) {
	dbConsts := DBConsts{
		"starter",    // Host
		8529,         // Port
		"checker_db", // DB
		"txn_g",      // TxnGraph
		"evt_g",      // EvtGraph
		"txn",        // TxnNode
		"w_evt",      // WriteEvtNode
		"r_evt",      // ReadEvtNode
		"dep",        // TxnDepEdge
		"evt_dep",    // EvtDepEdge
	}
	ednFileName := fmt.Sprintf("../histories/rw-register/%s.edn", fileName)
	walFileName := fmt.Sprintf("../histories/rw-register/%s.log", fileName)
	prompt := fmt.Sprintf("Checking %s...", ednFileName)
	fmt.Println(prompt)

	historyBuffer, err := os.ReadFile(ednFileName)
	if err != nil {
		log.Fatalf("Cannot read edn file %s", ednFileName)
		t.Fail()
	}
	history, err := core.ParseHistoryRW(string(historyBuffer))
	if err != nil {
		log.Fatalf("Cannot parse edn file %s", ednFileName)
		t.Fail()
	}

	walBuffer, err := os.ReadFile(walFileName)
	if err != nil {
		log.Fatalf("Cannot read wal log %s", walFileName)
		t.Fail()
	}
	wal, err := ParseWAL(string(walBuffer))
	if err != nil {
		log.Fatalf("Cannot parse wal log %s", walFileName)
		t.Fail()
	}

	t1 := time.Now()
	db, metadata := ConstructGraph(txn.Opts{}, history, wal, dbConsts)
	t2 := time.Now()
	constructTime := t2.Sub(t1).Nanoseconds() / 1e6
	fmt.Printf("constructing graph: %d ms\n", constructTime)

	return db, metadata, dbConsts
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
		db, metadata, dbConsts := constructGraph(strconv.Itoa(d), t)
		avgCheckTimeV1 := Profile(db, dbConsts, metadata, CheckSERV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, metadata, CheckSERV2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, metadata, CheckSERV3, false)
		avgCheckTimePregel := Profile(db, dbConsts, nil, CheckSERPregel, false)

		fmt.Printf("checking serializability (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - pregel: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimePregel)
		printLine()
	}
}

func TestProfilingSI(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, metadata, dbConsts := constructGraph(strconv.Itoa(d), t)

		avgCheckTimeV1 := Profile(db, dbConsts, metadata, CheckSIV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, metadata, CheckSIV2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, metadata, CheckSIV3, false)

		fmt.Printf("checking snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3)
		printLine()
	}
}

func TestProfilingPSI(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, metadata, dbConsts := constructGraph(strconv.Itoa(d), t)

		avgCheckTimeV1 := Profile(db, dbConsts, metadata, CheckPSIV1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, metadata, CheckPSIV2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, metadata, CheckPSIV3, false)

		fmt.Printf("checking parallel snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3)
		printLine()
	}
}

// go test -v -timeout 30s -run ^TestRWRegisterSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterSER(t *testing.T) {
	db, metadata, dbConsts := constructGraph("rw-register", t)
	valid := CheckSERV3(db, dbConsts, metadata, true)
	if !valid {
		fmt.Println("Not Serializable!")
	}
}

func TestRWRegisterSERPregel(t *testing.T) {
	db, _, dbConsts := constructGraph("rw-register", t)
	CheckSERPregel(db, dbConsts, nil, true)
}

// go test -v -timeout 30s -run ^TestRWRegisterSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterSI(t *testing.T) {
	db, metadata, dbConsts := constructGraph("rw-register", t)
	valid := CheckSIV3(db, dbConsts, metadata, true)
	if !valid {
		fmt.Println("Not Serializable!")
	}
}

// go test -v -timeout 30s -run ^TestRWRegisterPSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterPSI(t *testing.T) {
	db, metadata, dbConsts := constructGraph("rw-register", t)
	valid := CheckPSIV2(db, dbConsts, metadata, true)
	if !valid {
		fmt.Println("Not Serializable!")
	}
}
