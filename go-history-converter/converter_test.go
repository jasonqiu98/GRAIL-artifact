package gohistoryconverter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
)

func testConverter(path string, rw bool, t *testing.T) {
	pathList := strings.Split(path, string(os.PathSeparator))
	fileName := pathList[len(pathList)-1]
	newFileName := strings.Replace(fileName, ".edn", ".txt", 1)
	var destFolder string
	if rw {
		destFolder = "rw-register"
	} else {
		destFolder = "list-append"
	}
	newPath := filepath.Join(destFolder, newFileName)
	prompt := fmt.Sprintf("Converting %s to %s...", path, newPath)

	log.Println(prompt)
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read edn file %s", path)
		t.Fail()
	}
	var history core.History
	if rw {
		history, err = core.ParseHistoryRW(string(content))
		if err != nil {
			log.Fatalf("Cannot parse edn file %s", path)
			t.Fail()
		}
	} else {
		history, err = core.ParseHistory(string(content))
		if err != nil {
			log.Fatalf("Cannot parse edn file %s", path)
			t.Fail()
		}
	}

	f, err := os.Create(newPath)
	if err != nil {
		log.Println(err)
		t.Fail()
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	Converter(history, w)
	w.Flush()
}

func TestConverterListAppend(t *testing.T) {
	path := "../go-graph-checker/histories/list-append/10.edn"
	testConverter(path, false, t)
}

func TestConverterRWRegister(t *testing.T) {
	path := "../go-graph-checker/histories/rw-register/10.edn"
	testConverter(path, true, t)
}

func TestConvertsAll(t *testing.T) {
	for i := 1; i <= 20; i++ {
		{
			path := fmt.Sprintf("../go-graph-checker/histories/replication/rate/%v.edn", i*10)
			testConverter(path, false, t)
		}

		// {
		// 	path := fmt.Sprintf("../go-graph-checker/histories/rw-register/%v.edn", i*10)
		// 	testConverter(path, true, t)
		// }
	}
}
