package gohistoryconverter

import (
	"bufio"
	"fmt"
	"reflect"

	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
)

func preProcessHistory(history core.History) core.History {
	history = core.FilterOutNemesisHistory(history)
	history.AttachIndexIfNoExists()
	return history
}

/*
<r/w>(<key>,<value>,<sessionId>,<txnId>)
*/
func Converter(history core.History, w *bufio.Writer) {
	history = preProcessHistory(history)
	okHistory := core.FilterOkHistory(history)
	for _, t := range okHistory {
		sessionId := t.Process
		txnId := t.Index
		for _, op := range *t.Value {
			key, value := op.GetKey(), op.GetValue()
			if op.IsRead() {
				if value == nil {
					line := fmt.Sprintf("r(%v,%v,%v,%v)\n", key, 0, sessionId, txnId)
					w.WriteString(line)
					continue
				}

				if reflect.TypeOf(value).Kind() == reflect.Slice {
					valueSlice := reflect.ValueOf(value)
					if valueSlice.Len() == 0 {
						line := fmt.Sprintf("r(%v,%v,%v,%v)\n", key, 0, sessionId, txnId)
						w.WriteString(line)
						continue
					}
					lastValue := valueSlice.Index(valueSlice.Len() - 1)
					line := fmt.Sprintf("r(%v,%v,%v,%v)\n", key, lastValue, sessionId, txnId)
					w.WriteString(line)
				} else {
					line := fmt.Sprintf("r(%v,%v,%v,%v)\n", key, value, sessionId, txnId)
					w.WriteString(line)
				}
			} else {
				line := fmt.Sprintf("w(%v,%v,%v,%v)\n", key, value, sessionId, txnId)
				w.WriteString(line)
			}
		}
	}
}
