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
	"github.com/stretchr/testify/require"
)

func TestCheckExample(t *testing.T) {
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
	ednFileName := "../histories/anti-patterns/lost-update-2/history.edn"
	prompt := fmt.Sprintf("Checking %s...", ednFileName)
	log.Println(prompt)
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
	db, txnIds, g1 := ConstructGraph(txn.Opts{}, history, dbConsts)
	require.Equal(t, g1.G1a, false)
	require.Equal(t, g1.G1b, false)
	t2 := time.Now()
	constructTime := t2.Sub(t1).Nanoseconds() / 1e6
	fmt.Printf("constructing graph: %d ms\n", constructTime)

	{
		valid, cycle := CheckSERV3(db, dbConsts, txnIds, true)
		if !valid {
			log.Println("Not Serializable!")
			PlotCycle(history, cycle, "../images", "la-ser", false)
		} else {
			log.Println("Anti-Patterns for Serializable Not Detected.")
		}
	}

	{
		valid, cycle := CheckSIV3(db, dbConsts, txnIds, true)
		if !valid {
			log.Println("Not Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "la-si", false)
		} else {
			log.Println("Anti-Patterns for Snapshot Isolation Not Detected.")
		}
	}

	{
		valid, cycle := CheckPSIV3(db, dbConsts, txnIds, true)
		if !valid {
			log.Println("Not Parallel Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "la-psi", false)
		} else {
			log.Println("Anti-Patterns for Parallel Snapshot Isolation Not Detected.")
		}
	}

	{
		valid, cycle := CheckPL2V3(db, dbConsts, txnIds, true)
		if !valid {
			log.Println("Not PL-2!")
			PlotCycle(history, cycle, "../images", "la-pl2", false)
		} else {
			log.Println("Anti-Patterns for PL-2 Not Detected.")
		}
	}

	{
		valid, cycle := CheckPL1V3(db, dbConsts, txnIds, true)
		if !valid {
			log.Println("Not PL-1!")
			PlotCycle(history, cycle, "../images", "la-pl1", false)
		} else {
			log.Println("Anti-Patterns for PL-1 Not Detected.")
		}
	}
}

func printLine() {
	log.Println("-----------------------------------")
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
	ednFileName := fmt.Sprintf("../histories/list-append-scalability-rate/%s.edn", fileName)
	prompt := fmt.Sprintf("Checking %s...", ednFileName)
	log.Println(prompt)
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
	db, txnIds, g1 := ConstructGraph(txn.Opts{}, history, dbConsts)
	require.Equal(t, g1.G1a, false)
	require.Equal(t, g1.G1b, false)
	t2 := time.Now()
	constructTime := t2.Sub(t1).Nanoseconds() / 1e6
	fmt.Printf("constructing graph: %d ms\n", constructTime)

	return db, txnIds, dbConsts, history
}

func mustParseOp(opString string) core.Op {
	op, err := core.ParseOp(opString)
	if err != nil {
		log.Fatalf("expect no error, got %v", err)
	}
	return op
}

func TestProfilingScalability(t *testing.T) {
	var runtime [][]int64
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)
		var cur []int64
		{
			avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckSERV1, false)
			avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckSERV2, false)
			avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckSERV3, false)
			avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckSERV4, false)
			avgCheckTimePregel := Profile(db, dbConsts, nil, CheckSERPregel, false)
			cur = append(cur, avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4, avgCheckTimePregel)
		}

		{
			avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckSIV1, false)
			avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckSIV2, false)
			avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckSIV3, false)
			avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckSIV4, false)
			cur = append(cur, avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		}

		{
			avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckPSIV1, false)
			avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckPSIV2, false)
			avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckPSIV3, false)
			avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckPSIV4, false)
			cur = append(cur, avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		}

		{
			avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckPL2V1, false)
			avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckPL2V2, false)
			avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckPL2V3, false)
			avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckPL2V4, false)
			cur = append(cur, avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		}

		{
			avgCheckTimeV1 := Profile(db, dbConsts, txnIds, CheckPL1V1, false)
			avgCheckTimeV2 := Profile(db, dbConsts, txnIds, CheckPL1V2, false)
			avgCheckTimeV3 := Profile(db, dbConsts, txnIds, CheckPL1V3, false)
			avgCheckTimeV4 := Profile(db, dbConsts, txnIds, CheckPL1V4, false)
			cur = append(cur, avgCheckTimeV1, avgCheckTimeV2, avgCheckTimeV3, avgCheckTimeV4)
		}
		runtime = append(runtime, cur)
	}
	for _, row := range runtime {
		for _, col := range row {
			fmt.Print(col)
			fmt.Print("\t")
		}
		fmt.Println()
	}
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

func TestCorrectness(t *testing.T) {
	var res [][]bool
	for i := 1; i <= 30; i++ {
		var cur []bool
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(i), t)
		valid, _ := CheckSERV1(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckSERV2(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckSERV3(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckSERV4(db, dbConsts, txnIds, false)
		cur = append(cur, valid)

		valid, _ = CheckSIV1(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckSIV2(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckSIV3(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckSIV4(db, dbConsts, txnIds, false)
		cur = append(cur, valid)

		valid, _ = CheckPSIV1(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPSIV2(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPSIV3(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPSIV4(db, dbConsts, txnIds, false)
		cur = append(cur, valid)

		valid, _ = CheckPL2V1(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPL2V2(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPL2V3(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPL2V4(db, dbConsts, txnIds, false)
		cur = append(cur, valid)

		valid, _ = CheckPL1V1(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPL1V2(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPL1V3(db, dbConsts, txnIds, false)
		cur = append(cur, valid)
		valid, _ = CheckPL1V4(db, dbConsts, txnIds, false)
		cur = append(cur, valid)

		res = append(res, cur)
	}

	for _, row := range res {
		for _, col := range row {
			fmt.Print(col)
			fmt.Print(" ")
		}
		fmt.Println()
	}
}

// go test -v -timeout 30s -run ^TestListAppendSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendSER(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("20", t)
	valid, cycle := CheckSERV1(db, dbConsts, txnIds, true)
	if !valid {
		log.Println("Not Serializable!")
		PlotCycle(history, cycle, "../images", "la-ser", true)
	} else {
		log.Println("Anti-Patterns Not Detected.")
	}
}

func TestListAppendSERPregel(t *testing.T) {
	db, _, dbConsts, _ := constructArangoGraph("list-append", t)
	CheckSERPregel(db, dbConsts, nil, true)
}

// go test -v -timeout 30s -run ^TestListAppendSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("list-append", t)
	valid, cycle := CheckSIV4(db, dbConsts, txnIds, true)
	if !valid {
		log.Println("Not Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "la-si", true)
	} else {
		log.Println("Anti-Patterns Not Detected.")
	}
}

// go test -v -timeout 30s -run ^TestListAppendPSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendPSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("list-append", t)
	valid, cycle := CheckPSIV2(db, dbConsts, txnIds, true)
	if !valid {
		log.Println("Not Parallel Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "la-psi", true)
	} else {
		log.Println("Anti-Patterns Not Detected.")
	}
}

// Tests for correctness, following TDD principles

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
	dbConsts := DBConsts{"starter", 8529, "checker_db", "txn_g", "evt_g", "txn", "a_evt", "r_evt", "dep", "evt_dep"}

	{
		// G0 (write cycles) ~ violates PL-1
		log.Println("Checking G0...")
		t1 := mustParseOp(`{:type :ok, :value [[:append x 1] [:append y 1]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:append x 2] [:append y 2]]}`)
		t3 := mustParseOp(`{:type :ok, :value [[:r x [1 2]] [:r y [2 1]]]}`)
		h := []core.Op{t1, t2, t3}

		// checking G0 doesn't require G1b and G1c
		// however, for simplicity, we just keep the checks
		db, txnIds, _ := ConstructGraph(txn.Opts{}, h, dbConsts)

		// expect the result to be false
		testPL1(t, h, db, dbConsts, txnIds, false)
	}

	{
		// G1c (circular information flow) ~ violates PL-2 but not PL-1
		log.Println("Checking G1c...")
		t1 := mustParseOp(`{:type :ok, :value [[:append x 1] [:r y [1]]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:append x 2] [:append y 1]]}`)
		t3 := mustParseOp(`{:type :ok, :value [[:r x [1 2]] [:r y [1]]]}`)
		h := []core.Op{t1, t2, t3}

		db, txnIds, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

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
		t1 := mustParseOp(`{:type :ok, :value [[:r x []] [:append x 1] [:append x 2]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:r x [1]]]}`)
		h := []core.Op{t1, t2}

		_, _, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

		require.Equal(t, g1.G1b, true)
	}

	{
		// a G-single case (single anti-dependency cycle)
		// in Adya's PhD thesis
		// proscribed by PL-2+ and above, but not PL-2
		log.Println("Checking G-single...")
		h := []core.Op{
			mustParseOp(`{:type :ok, :value [[:append 1 1] [:append 2 1]]}`),
			mustParseOp(`{:type :ok, :value [[:append 1 2] [:append 2 2]]}`),
			mustParseOp(`{:type :ok, :value [[:r 1 [1 2]] [:r 2 [1]]]}`),
		}

		db, txnIds, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

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
		t1 := mustParseOp(`{:type :ok, :value [[:r x []] [:append x 1]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:r x []] [:r x [1]]]}`)
		h := []core.Op{t1, t2}

		db, txnIds, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, false)
	}

	{
		// lost update ~ violates SER, SI and PSI
		log.Println("Checking lost update...")
		t1 := mustParseOp(`{:type :ok, :value [[:r x []] [:append x 1]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:r x []] [:append x 2]]}`)
		t3 := mustParseOp(`{:type :ok, :value [[:r x [2]]]}`)
		h := []core.Op{t1, t2, t3}

		db, txnIds, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, false)
	}

	// conflicting fork ~ violates SER, SI and PSI
	// doesn't satisfy the assumption that each object only has one append per value
	// will throw "non-recoverable" anomaly
	// {
	// 	log.Println("Checking conflicting fork...")
	// 	t1 := mustParseOp(`{:type :ok, :value [[:r x []] [:append x 1]]}`)
	// 	t2 := mustParseOp(`{:type :ok, :value [[:append x 1]]}`)
	// 	t3 := mustParseOp(`{:type :ok, :value [[:r x [1 2]]]}`)
	// 	h := []core.Op{t1, t2, t3}

	// 	ConstructGraph(txn.Opts{}, h, dbConsts)
	// }

	{
		// long fork ~ violates SER and SI but not PSI
		log.Println("Checking long fork...")
		t1 := mustParseOp(`{:type :ok, :value [[:r x []] [:r y []] [:append x 1]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:r x [1]] [:r y []]]}`)
		t3 := mustParseOp(`{:type :ok, :value [[:r x []] [:r y []] [:append y 1]]}`)
		t4 := mustParseOp(`{:type :ok, :value [[:r x []] [:r y [1]]]}`)
		h := []core.Op{t1, t2, t3, t4}

		db, txnIds, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, false)
		testPSI(t, h, db, dbConsts, txnIds, true)
	}

	{
		// write skew / short fork ~ violates SER but not SI nor PSI
		log.Println("Checking write skew...")
		t1 := mustParseOp(`{:type :ok, :value [[:r x []] [:r y []] [:append x 1]]}`)
		t2 := mustParseOp(`{:type :ok, :value [[:r x []] [:r y []] [:append y 1]]}`)
		h := []core.Op{t1, t2}

		db, txnIds, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

		require.Equal(t, g1.G1a, false)
		require.Equal(t, g1.G1b, false)

		testSER(t, h, db, dbConsts, txnIds, false)
		testSI(t, h, db, dbConsts, txnIds, true)
		testPSI(t, h, db, dbConsts, txnIds, true)
	}

}

/*
G1a aborted read
*/
func TestG1aCases(t *testing.T) {
	dbConsts := DBConsts{"starter", 8529, "checker_db", "txn_g", "evt_g", "txn", "a_evt", "r_evt", "dep", "evt_dep"}

	t1 := mustParseOp(`{:type :fail, :value [[:append x 1]]}`)
	t2 := mustParseOp(`{:type :ok, :value [[:r x [1]] [:append x 2]]}`)
	t3 := mustParseOp(`{:type :ok, :value [[:r x [1 2]] [:r y [3]]]}`)

	h := []core.Op{t2, t3, t1}

	_, _, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1a, true) // G1a detected
}

/*
G1b intermediate read
*/
func TestG1bCases(t *testing.T) {
	dbConsts := DBConsts{"starter", 8529, "checker_db", "txn_g", "evt_g", "txn", "a_evt", "r_evt", "dep", "evt_dep"}

	h := []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1] [:append x 2]]}`),
		mustParseOp(`{:type :ok, :value [[:r x [1]]]}`),
	}

	_, _, g1 := ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, true) // G1b detected

	h = []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1]]}`),
		mustParseOp(`{:type :ok, :value [[:append x 2] [:append x 3] [:r x [1 2]]]}`),
	}

	_, _, g1 = ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, true) // G1b detected

	h = []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1]]}`),
		mustParseOp(`{:type :ok, :value [[:append x 2] [:append x 3] [:r x [1 2]]]}`),
		mustParseOp(`{:type :ok, :value [[:r x [1 2 3]]]}`),
	}

	_, _, g1 = ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, true) // G1b detected

	h = []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1]]}`),
		mustParseOp(`{:type :ok, :value [[:append x 2] [:append x 3] [:r x [1 2]]]}`),
		mustParseOp(`{:type :ok, :value [[:r x [1 2]]]}`),
	}

	_, _, g1 = ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, true) // G1b detected

	h = []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1]]}`),
		mustParseOp(`{:type :ok, :value [[:append x 2] [:r x [1 2]] [:append x 3]]}`),
	}

	_, _, g1 = ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, false) // G1b not detected

	h = []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1] [:append x 2]]}`),
		mustParseOp(`{:type :ok, :value [[:append x 3]]}`),
		mustParseOp(`{:type :ok, :value [[:r x [1 2 3]]]}`),
	}

	_, _, g1 = ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, false) // G1b not detected

	h = []core.Op{
		mustParseOp(`{:type :ok, :value [[:append x 1] [:append x 2]]}`),
		mustParseOp(`{:type :ok, :value [[:append x 3]]}`),
		mustParseOp(`{:type :ok, :value [[:r x [1 2]]]}`),
	}

	_, _, g1 = ConstructGraph(txn.Opts{}, h, dbConsts)

	require.Equal(t, g1.G1b, false) // G1b not detected

}
