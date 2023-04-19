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
	"github.com/stretchr/testify/require"
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
	db, txnIds, g1 := ConstructGraph(txn.Opts{}, history, wal, dbConsts)
	require.Equal(t, g1.G1a, false)
	require.Equal(t, g1.G1b, false)
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
		avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckSERV4, false)
		avgCheckTimePregel := Profile(db, dbConsts, nil, CheckSERPregel, false)

		fmt.Printf("checking serializability (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - v4: %d ms\n - pregel: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4, avgCheckTimePregel)
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
		avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckSIV4, false)

		fmt.Printf("checking snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - v4: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
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
		avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckPSIV4, false)

		fmt.Printf("checking parallel snapshot isolation (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - v4: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		printLine()
	}
}

func TestProfilingPL2(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)

		avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckPL2V1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckPL2V2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckPL2V3, false)
		avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckPL2V4, false)

		fmt.Printf("checking PL-2 (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - v4: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		printLine()
	}
}

func TestProfilingPL1(t *testing.T) {
	printLine()
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)

		avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckPL1V1, false)
		avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckPL1V2, false)
		avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckPL1V3, false)
		avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckPL1V4, false)

		fmt.Printf("checking PL-1 (on avg.):\n - v1: %d ms\n - v2: %d ms\n - v3: %d ms\n - v4: %d ms\n",
			avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		printLine()
	}
}

// go test -v -timeout 30s -run ^TestRWRegisterSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterSER(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("170", t)
	valid, cycle := CheckSERV1(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Serializable!")
		PlotCycle(history, cycle, "../images", "rw-ser", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

func TestRWRegisterSERPregel(t *testing.T) {
	db, _, dbConsts, _ := constructArangoGraph("rw-register", t)
	CheckSERPregel(db, dbConsts, nil, true)
}

// go test -v -timeout 30s -run ^TestRWRegisterSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("rw-register", t)
	valid, cycle := CheckSIV3(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "rw-si", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

// go test -v -timeout 30s -run ^TestRWRegisterPSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterPSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("rw-register", t)
	valid, cycle := CheckPSIV3(db, dbConsts, txnIds, true)
	if !valid {
		fmt.Println("Not Parallel Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "rw-psi", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

/// Tests for correctness, following TDD principles

func testPL1(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	valid, _ := CheckPL1V1(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPL1V2(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPL1V3(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPL1V4(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)
}

func testPL2(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	valid, _ := CheckPL2V1(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPL2V2(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPL2V3(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPL2V4(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)
}

func testPSI(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	valid, _ := CheckPSIV1(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPSIV2(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPSIV3(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckPSIV4(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)
}

func testSI(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	valid, _ := CheckSIV1(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSIV2(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSIV3(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSIV4(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)
}

func testSER(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	valid, _ := CheckSERV1(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSERV2(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSERV3(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSERV4(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)

	valid, _ = CheckSERPregel(db, dbConsts, txnIds, false)
	require.Equal(t, expected, valid)
}
func TestChecker(t *testing.T) {
	{
		// G0 (write cycles) ~ violates PL-1
		log.Println("Checking G0...")
		db, txnIds, dbConsts, h, _ := testConstructArangoGraph("g0", t)
		// checking G0 doesn't require checking of G1
		// expect the result to be false
		testPL1(t, h, db, dbConsts, txnIds, false)
	}

	{
		// G1c (circular information flow) ~ violates PL-2 but not PL-1
		log.Println("Checking G1c...")
		db, txnIds, dbConsts, h, g1 := testConstructArangoGraph("g1c", t)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		// expect the result to be false
		testPL2(t, h, db, dbConsts, txnIds, false)
		testPL1(t, h, db, dbConsts, txnIds, true)
	}

	// from paper: Transactional storage for geo-replicated systems

	{
		// dirty read (G1b) ~ violates SER, SI and PSI
		log.Println("Checking dirty read...")
		_, _, _, _, g1 := testConstructArangoGraph("dirty-read", t)

		require.Equal(t, g1.G1b, true)
	}

	{
		// a G-single case (single anti-dependency cycle)
		// in Adya's PhD thesis
		// proscribed by PL-2+ and above, but not PL-2
		log.Println("Checking G-single...")
		db, txnIds, dbConsts, h, g1 := testConstructArangoGraph("g-single", t)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testPL2(t, h, db, dbConsts, txnIds, true) // PL-2 not violated

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, false)
	}

	{
		// non-repeatable read ~ violates SER, SI and PSI
		log.Println("Checking non-repeatable read...")
		db, txnIds, dbConsts, h, g1 := testConstructArangoGraph("non-repeatable-read", t)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, false)
	}

	{
		// lost update ~ violates SER, SI and PSI
		log.Println("Checking lost update...")
		db, txnIds, dbConsts, h, g1 := testConstructArangoGraph("lost-update", t)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, false)
	}

	{
		// long fork ~ violates SER and SI but not PSI
		log.Println("Checking long fork...")
		db, txnIds, dbConsts, h, g1 := testConstructArangoGraph("long-fork", t)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, true)
	}

	{
		// write skew / short fork ~ violates SER but not SI nor PSI
		log.Println("Checking write skew...")
		db, txnIds, dbConsts, h, g1 := testConstructArangoGraph("write-skew", t)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, true)
		testPSI(t, h, db, dbConsts, txnIds, true)
	}

}

func testConstructArangoGraph(fileName string, t *testing.T) (driver.Database, []int, DBConsts, core.History, G1Anomalies) {
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
	ednFileName := fmt.Sprintf("../histories/rw-register-test/%s.edn", fileName)
	walFileName := fmt.Sprintf("../histories/rw-register-test/%s.log", fileName)
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

	db, txnIds, g1 := ConstructGraph(txn.Opts{}, history, wal, dbConsts)
	return db, txnIds, dbConsts, history, g1
}

/*
G1a aborted read
*/
func TestG1aCases(t *testing.T) {
	_, _, _, _, g1 := testConstructArangoGraph("g1a", t)
	require.Equal(t, g1.G1a, true)
}

/*
G1b intermediate read
*/
func TestG1bCases(t *testing.T) {
	_, _, _, _, g1 := testConstructArangoGraph("g1b-1", t)
	require.Equal(t, g1.G1b, true)

	_, _, _, _, g1 = testConstructArangoGraph("g1b-2", t)
	require.Equal(t, g1.G1b, false)
}
