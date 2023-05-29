package listappend

import (
	"context"
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
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "ser", "sp")
		if !valid {
			log.Println("Not Serializable!")
			PlotCycle(history, cycle, "../images", "la-ser", false)
		} else {
			log.Println("Anti-Patterns for Serializable Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "si", "sp")
		if !valid {
			log.Println("Not Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "la-si", false)
		} else {
			log.Println("Anti-Patterns for Snapshot Isolation Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "psi", "sp")
		if !valid {
			log.Println("Not Parallel Snapshot Isolation!")
			PlotCycle(history, cycle, "../images", "la-psi", false)
		} else {
			log.Println("Anti-Patterns for Parallel Snapshot Isolation Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "pl-2", "sp")
		if !valid {
			log.Println("Not PL-2!")
			PlotCycle(history, cycle, "../images", "la-pl2", false)
		} else {
			log.Println("Anti-Patterns for PL-2 Not Detected.")
		}
	}

	{
		valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "pl-1", "sp")
		if !valid {
			log.Println("Not PL-1!")
			PlotCycle(history, cycle, "../images", "la-pl1", false)
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
		"a_evt",      // AppendEvtNode
		"r_evt",      // ReadEvtNode
		"dep",        // TxnDepEdge
		"evt_dep",    // EvtDepEdge
	}
	ednFileName := fmt.Sprintf("../histories/replication/histories-30s/%s.edn", fileName)
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
			t1 := Profile(db, dbConsts, txnIds, CheckSERSV, false)
			// t2 := Profile(db, dbConsts, txnIds, CheckSERSVFilter, false)
			// t3 := Profile(db, dbConsts, txnIds, CheckSERSP, false)
			// tp := Profile(db, dbConsts, nil, CheckSERPregel, false)
			// cur = append(cur, t1, t2, t3, tp)
			cur = append(cur, t1)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckSISV, false)
			// t3 := Profile(db, dbConsts, txnIds, CheckSISP, false)
			// cur = append(cur, t1, t3)
			cur = append(cur, t1)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckPSISV, false)
			// t3 := Profile(db, dbConsts, txnIds, CheckPSISP, false)
			// cur = append(cur, t1, t3)
			cur = append(cur, t1)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckPL2SV, false)
			// t3 := Profile(db, dbConsts, txnIds, CheckPL2SP, false)
			// cur = append(cur, t1, t3)
			cur = append(cur, t1)
		}

		{
			t1 := Profile(db, dbConsts, txnIds, CheckPL1SV, false)
			// t3 := Profile(db, dbConsts, txnIds, CheckPL1SP, false)
			// cur = append(cur, t1, t3)
			cur = append(cur, t1)
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
		for _, level := range []string{"ser", "si", "psi"} {
			for _, mode := range []string{"sv"} {
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

// go test -v -timeout 30s -run ^TestListAppendSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/list_append
func TestListAppendSER(t *testing.T) {
	db, txnIds, dbConsts, history := constructArangoGraph("70", t)
	valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "ser", "sv")
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
	db, txnIds, dbConsts, history := constructArangoGraph("150", t)
	valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "si", "sv")
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
	valid, cycle := IsolationLevelChecker(db, dbConsts, txnIds, true, "psi", "sv")
	if !valid {
		log.Println("Not Parallel Snapshot Isolation!")
		PlotCycle(history, cycle, "../images", "la-psi", true)
	} else {
		log.Println("Anti-Patterns Not Detected.")
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
