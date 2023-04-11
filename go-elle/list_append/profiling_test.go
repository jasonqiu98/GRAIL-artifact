package listappend

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func TestListAppendSER(t *testing.T) {
	content, err := ioutil.ReadFile("../histories/list-append.edn")
	if err != nil {
		t.Fail()
	}
	history, err := core.ParseHistory(string(content))
	if err != nil {
		t.Fail()
	}
	t1 := time.Now()
	result := Check(txn.Opts{
		ConsistencyModels: []core.ConsistencyModelName{"serializable"},
	}, history)
	t2 := time.Now()
	if !result.Valid {
		fmt.Println("Not Serializable!")
	}
	stage := t2.Sub(t1).Nanoseconds() / 1e6
	fmt.Printf("checking serializability: %dms\n", stage)
}
