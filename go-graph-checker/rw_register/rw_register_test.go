package rwregister

import (
	"fmt"
	"os"
	"testing"

	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

// go test -v -timeout 30s -run ^TestRWRegisterSER$ github.com/jasonqiu98/anti-pattern-graph-checker-single/go-graph-checker/rw_register
func TestRWRegisterSER(t *testing.T) {
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
	historyBuffer, err := os.ReadFile("../histories/rw-register.edn")
	if err != nil {
		t.Fail()
	}
	history, err := core.ParseHistoryRW(string(historyBuffer))
	if err != nil {
		t.Fail()
	}

	walBuffer, err := os.ReadFile("../histories/rw-register.log")
	if err != nil {
		t.Fail()
	}
	wal, err := ParseWAL(string(walBuffer))
	if err != nil {
		t.Fail()
	}

	db, metadata := ConstructGraph(txn.Opts{}, history, wal, dbConsts)
	valid := CheckSERV1(db, dbConsts, metadata, true)
	if !valid {
		fmt.Println("Not Serializable!")
	}
	// db, _ := ConstructGraph(txn.Opts{}, history, wal, dbConsts)
	// CheckSERPregel(db, dbConsts, nil, true)
}
