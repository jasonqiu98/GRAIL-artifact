package rwregister

import (
	"fmt"
	"os"
	"testing"
)

// go test -v -timeout 30s -run ^TestParseWAL$ github.com/grail/anti-pattern-graph-checker-single/go-graph-checker/rw_register
func TestParseWAL(t *testing.T) {
	fileName := "../histories/rw-register.log"
	prompt := fmt.Sprintf("Checking %s...", fileName)
	fmt.Println(prompt)

	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fail()
	}
	wal, err := ParseWAL(string(content))
	if err != nil {
		t.Fail()
	}

	_ = wal
}

// go test -v -timeout 30s -run ^TestFilterWALWrite$ github.com/grail/anti-pattern-graph-checker-single/go-graph-checker/rw_register
func TestWALWriteMap(t *testing.T) {
	fileName := "../histories/rw-register.log"
	prompt := fmt.Sprintf("Checking %s...", fileName)
	fmt.Println(prompt)

	content, err := os.ReadFile(fileName)
	if err != nil {
		t.Fail()
	}
	wal, err := ParseWAL(string(content))
	if err != nil {
		t.Fail()
	}
	// Data: "_id" string, "_key" string, "_rev" string, "rwAttr" int32
	// writeMap and writeInvMap
	wm := ConstructWALWriteMap(wal, "rwAttr")
	_ = wm
}
