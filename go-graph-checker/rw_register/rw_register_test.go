package rwregister

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/grail/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/grail/anti-pattern-graph-checker-single/go-elle/txn"
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
		"w_evt",      // WriteEvtNode
		"r_evt",      // ReadEvtNode
		"dep",        // TxnDepEdge
		"evt_dep",    // EvtDepEdge
	}
	ednFileName := fmt.Sprintf("../histories/rw-register/%s.edn", "20")
	walFileName := fmt.Sprintf("../histories/rw-register/%s.log", "20")
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

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "ser", "sp")
		if !valid {
			log.Println("Not Serializable!")
			PlotCycle(history, cycle, "../images", "rw-ser", false)
		} else {
			log.Println("Anti-Patterns for Serializable Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "si", "sp")
		if !valid {
			log.Println("Not Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "rw-si", false)
		} else {
			log.Println("Anti-Patterns for Snapshot Isolation Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "psi", "sp")
		if !valid {
			log.Println("Not Parallel Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "rw-psi", false)
		} else {
			log.Println("Anti-Patterns for Parallel Snapshot Isolation Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "pl-2", "sp")
		if !valid {
			log.Println("Not PL-2!")
			PlotCycle(history, cycle, "../images", "rw-pl2", false)
		} else {
			log.Println("Anti-Patterns for PL-2 Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "pl-1", "sp")
		if !valid {
			log.Println("Not PL-1!")
			PlotCycle(history, cycle, "../images", "rw-pl1", false)
		} else {
			log.Println("Anti-Patterns for PL-1 Not Detected.")
		}
	}
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

func TestProfilingScalability(t *testing.T) {
	var runtime [][]int64
	for d := 10; d <= 200; d += 10 {
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(d), t)
		var cur []int64
		{
			t1 := Profile(db, dbConsts, txnIds, CheckSERSV, false)
			t2 := Profile(db, dbConsts, txnIds, CheckSERSVFilter, false)
			t3 := Profile(db, dbConsts, txnIds, CheckSERSP, false)
			tp := Profile(db, dbConsts, nil, CheckSERPregel, false)
			cur = append(cur, t1, t2, t3, tp)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckSISV, false)
			t2 := Profile(db, dbConsts, txnIds, CheckSISP, false)
			cur = append(cur, t1, t2)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckPSISV, false)
			t2 := Profile(db, dbConsts, txnIds, CheckPSISP, false)
			cur = append(cur, t1, t2)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckPL2SV, false)
			t2 := Profile(db, dbConsts, txnIds, CheckPL2SP, false)
			cur = append(cur, t1, t2)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckPL1SV, false)
			t2 := Profile(db, dbConsts, txnIds, CheckPL1SP, false)
			cur = append(cur, t1, t2)
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

func TestCountCycles(t *testing.T) {
	for i := 1; i <= 20; i++ {
		db, _, _, _ := constructArangoGraph(strconv.Itoa(i*10), t)

		cursor, err := db.Query(context.Background(), "RETURN LENGTH(FOR edge IN dep FOR p IN OUTBOUND K_SHORTEST_PATHS edge._to TO edge._from GRAPH txn_g RETURN {edges: UNSHIFT(p.edges, edge), vertices: UNSHIFT(p.vertices, p.vertices[LENGTH(p.vertices) - 1])})", nil)
		// cursor, err := db.Query(context.Background(), "RETURN LENGTH(FOR edge IN dep FOR v, e IN OUTBOUND SHORTEST_PATH edge._to TO edge._from GRAPH txn_g RETURN [edge, e])", nil)
		// cursor, err := db.Query(context.Background(), "RETURN LENGTH(FOR start IN txn FOR vertex, edge, path IN 2..5 OUTBOUND start._id GRAPH txn_g FILTER edge._to == start._id RETURN path.edges)", nil)

		if err != nil {
			log.Fatalf("Failed to count: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle interface{}
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				fmt.Println(cycle)
			}
		}

	}
}

func TestCountVerticesEdges(t *testing.T) {
	for i := 1; i <= 30; i++ {
		db, _, _, _ := constructArangoGraph(strconv.Itoa(i), t)

		cursor, err := db.Query(context.Background(), "RETURN [(RETURN LENGTH(txn)), (RETURN LENGTH(dep))]", nil)
		if err != nil {
			log.Fatalf("Failed to count: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle interface{}
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				fmt.Println(cycle)
			}
		}

	}
}

func TestCorrectness(t *testing.T) {
	var res [][]bool
	for i := 10; i <= 200; i += 10 {
		var cur []bool
		db, txnIds, dbConsts, _ := constructArangoGraph(strconv.Itoa(i), t)
		for _, level := range []string{"ser", "si", "psi", "pl-2", "pl-1"} {
			for _, mode := range []string{"sv", "sv-filter", "sp"} {
				valid, _ := IsolationLevelChecker(db, dbConsts, txnIds, false, level, mode)
				cur = append(cur, valid)
			}
		}
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

// go test -v -timeout 30s -run ^TestRWRegisterSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterSER(t *testing.T) {
	for i := 10; i <= 200; i += 10 {
		db, txnIds, dbConsts, history := constructArangoGraph(strconv.Itoa(i), t)
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "ser", "sv")
		if !valid {
			fmt.Println("Not Serializable!")
			PlotCycle(history, cycle, "../images", "rw-ser", true)
		} else {
			fmt.Println("Anti-Patterns Not Detected.")
		}
	}
}

func TestRWRegisterSERPregel(t *testing.T) {
	db, _, dbConsts, _ := constructArangoGraph("10", t)
	CheckSERPregel(db, dbConsts, nil, true)
}

// go test -v -timeout 30s -run ^TestRWRegisterSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterSI(t *testing.T) {
	for i := 10; i <= 200; i += 10 {
		db, txnIds, dbConsts, history := constructArangoGraph(strconv.Itoa(i), t)
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "si", "sv")
		if !valid {
			fmt.Println("Not Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "rw-si", true)
		} else {
			fmt.Println("Anti-Patterns Not Detected.")
		}
	}
}

// go test -v -timeout 30s -run ^TestRWRegisterPSI$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestRWRegisterPSI(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("10", t)
	valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "psi", "sv")
	if !valid {
		fmt.Println("Not Parallel Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "rw-psi", true)
	} else {
		fmt.Println("Anti-Patterns Not Detected.")
	}
}

// Tests for correctness, following TDD principles

func testPL1(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	for _, mode := range []string{"sv", "sv-filter", "sv-random", "sp", "sp-allcycles"} {
		valid, _ := IsolationLevelChecker(db, dbConsts, txnIds, false, "pl-1", mode)
		require.Equal(t, expected, valid)
	}
}

func testPL2(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	for _, mode := range []string{"sv", "sv-filter", "sv-random", "sp", "sp-allcycles"} {
		valid, _ := IsolationLevelChecker(db, dbConsts, txnIds, false, "pl-2", mode)
		require.Equal(t, expected, valid)
	}
}

func testPSI(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	for _, mode := range []string{"sv", "sv-filter", "sv-random", "sp", "sp-allcycles"} {
		valid, _ := IsolationLevelChecker(db, dbConsts, txnIds, false, "psi", mode)
		require.Equal(t, expected, valid)
	}
}

func testSI(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	for _, mode := range []string{"sv", "sv-filter", "sv-random", "sp", "sp-allcycles"} {
		valid, _ := IsolationLevelChecker(db, dbConsts, txnIds, false, "si", mode)
		require.Equal(t, expected, valid)
	}
}

func testSER(t *testing.T, h core.History, db driver.Database, dbConsts DBConsts, txnIds []int, expected bool) {
	for _, mode := range []string{"sv", "sv-filter", "sv-random", "sp", "sp-allcycles"} {
		valid, _ := IsolationLevelChecker(db, dbConsts, txnIds, false, "ser", mode)
		require.Equal(t, expected, valid)
	}
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
