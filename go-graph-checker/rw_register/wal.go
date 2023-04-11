package rwregister

import (
	"encoding/json"
	"strings"
)

type WALEntry struct {
	Tick string                 `json:"tick"`
	Type int32                  `json:"type"`
	CUID string                 `json:"cuid"`
	DB   string                 `json:"db"`
	TID  string                 `json:"tid"`
	Data map[string]interface{} `json:"data"`
}

/*
TODO: add more WAL operation types here (optional)
*/
const (
	WALTypeInsertDoc int32 = 2300
)

type WAL []WALEntry

/*
parse lines of WAL logs into an WAL array
*/
func ParseWAL(content string) (WAL, error) {
	var wal WAL
	for _, line := range strings.Split(content, "\n") {
		if line == "" {
			continue
		}
		entry, err := ParseWALEntry(strings.Trim(line, " "))
		if err != nil {
			return nil, err
		}
		wal = append(wal, entry)
	}
	return wal, nil
}

/*
read each WAL log entry into a json
*/
func ParseWALEntry(entryString string) (WALEntry, error) {
	var (
		empty WALEntry
		entry WALEntry
	)
	err := json.Unmarshal([]byte(entryString), &entry)
	if err != nil {
		return empty, err
	}
	return entry, nil
}

/*
WALWriteMap: {key1: [v11 v12 v13 ...], key2: [v21 v22 v23, ...], ...}
WALWriteInvMap: {key1: {v11:0 v12:1 v13:2 ...}, key2: {v21:0 v22:1 v23:2 ...}, ...}
*/

type WALWriteMap map[string][]int
type WALWriteInvMap map[string]map[int]int

/*
needs to specify
`attr` -- the attribute name of the register used in the database
*/
func ConstructWALWriteMap(wal WAL, attr string) WALWriteMap {
	wm := make(map[string][]int)
	for _, l := range wal {
		if l.Type == WALTypeInsertDoc {
			key := l.Data["_key"].(string)
			val := int(l.Data[attr].(float64))
			// init the entries for both maps
			if _, ok := wm[key]; !ok {
				wm[key] = []int{}
			}
			wm[key] = append(wm[key], val)
		}
	}
	return wm
}
