package rwregister

import (
	"fmt"
	"os"
	"testing"

	driver "github.com/arangodb/go-driver"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func Benchmark(b *testing.B) {
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
	i := 120
	ednFileName := fmt.Sprintf("../histories/rw-register/%d.edn", i)
	walFileName := fmt.Sprintf("../histories/rw-register/%d.log", i)
	historyBuffer, err := os.ReadFile(ednFileName)
	if err != nil {
		b.Fail()
	}
	history, err := core.ParseHistoryRW(string(historyBuffer))
	if err != nil {
		b.Fail()
	}

	walBuffer, err := os.ReadFile(walFileName)
	if err != nil {
		b.Fail()
	}
	wal, err := ParseWAL(string(walBuffer))
	if err != nil {
		b.Fail()
	}

	db, txnIds := ConstructGraph(txn.Opts{}, history, wal, dbConsts)

	// ignore the cost to construct the dependency graph
	b.ResetTimer()
	// only benchmark the checking process
	benchmarkWrapper(db, dbConsts, txnIds, CheckSIV1, false, b)
}

func benchmarkWrapper(db driver.Database, dbConsts DBConsts, txnIds []int,
	f func(driver.Database, DBConsts, []int, bool) (bool, []TxnDepEdge),
	output bool, b *testing.B) {
	for n := 0; n < b.N; n++ {
		f(db, dbConsts, txnIds, output)
	}
}
