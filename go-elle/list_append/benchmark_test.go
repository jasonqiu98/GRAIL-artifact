package listappend

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func BenchmarkListAppendSER(b *testing.B) {
	content, err := ioutil.ReadFile("../histories/list-append/list-append.edn")
	if err != nil {
		b.Fail()
	}
	history, err := core.ParseHistory(string(content))
	if err != nil {
		b.Fail()
	}

	// ignore the prep work
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		result := Check(txn.Opts{
			ConsistencyModels: []core.ConsistencyModelName{"serializable"},
		}, history)
		if !result.Valid {
			fmt.Println("Not Serializable!")
		}
	}
}

func BenchmarkListAppendSI(b *testing.B) {
	content, err := ioutil.ReadFile("../histories/list-append/list-append.edn")
	if err != nil {
		b.Fail()
	}
	history, err := core.ParseHistory(string(content))
	if err != nil {
		b.Fail()
	}

	// ignore the prep work
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		result := Check(txn.Opts{
			ConsistencyModels: []core.ConsistencyModelName{"snapshot-isolation"},
		}, history)
		if !result.Valid {
			fmt.Println("Not Snapshot Isolation!")
		}
	}
}

func BenchmarkListAppendPSI(b *testing.B) {
	content, err := ioutil.ReadFile("../histories/list-append/list-append.edn")
	if err != nil {
		b.Fail()
	}
	history, err := core.ParseHistory(string(content))
	if err != nil {
		b.Fail()
	}

	// ignore the prep work
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		result := Check(txn.Opts{
			ConsistencyModels: []core.ConsistencyModelName{"parallel-snapshot-isolation"},
		}, history)
		if !result.Valid {
			fmt.Println("Not Parallel Snapshot Isolation!")
		}
	}
}
