package rwregister

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
)

/*
Key: "i", reflecting the order
Index: index attached, coming with the history
*/
type TxnNode struct {
	Key string `json:"_key"`
}

/*
f(x, xi, a) -> (xj, r)
Key: "i,j": i is the order of txn; j is the order of evt
Key is the key for the evt in the checker database
Obj is the original key of the history
*/

type WriteEvt struct {
	Key   string `json:"_key"`
	Obj   string `json:"obj"`
	Arg   int    `json:"arg"`
	Index int    `json:"index"` // -1 means last write
}

type ReadEvt struct {
	Key string `json:"_key"`
	Obj string `json:"obj"`
	V   int    `json:"v"` // 0 ~ nil, >=1 ~ possible values
}

type DBConsts struct {
	Host         string
	Port         int
	DB           string
	TxnGraph     string
	EvtGraph     string
	TxnNode      string
	WriteEvtNode string
	ReadEvtNode  string
	TxnDepEdge   string
	EvtDepEdge   string
}

/*
borrowed from go-elle
*/
func preProcessHistory(history core.History) core.History {
	history = core.FilterOutNemesisHistory(history)
	history.AttachIndexIfNoExists()
	return history
}

/*
returns a client instance of ArangoDB
*/
func startClient(host string, port int) driver.Client {
	endpoint := fmt.Sprintf("http://%s:%d", host, port)
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{endpoint},
	})
	if err != nil {
		log.Fatalf("Failed to connect to the host %s at port %d: %v\n", host, port, err)
	}
	client, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		log.Fatalf("Failed to create a new client: %v\n", err)
	}
	return client
}

/*
get db if db already exists, otherwise create db
*/
func getOrCreateDB(client driver.Client, dbName string) driver.Database {
	if dbExists, _ := client.DatabaseExists(context.Background(), dbName); dbExists {
		log.Println("db exists already and will be dropped first...")
		db, err := client.Database(context.Background(), dbName)
		if err != nil {
			log.Fatalf("Failed to open existing database: %v\n", err)
		}
		// forcibly drop the database for a new checker
		db.Remove(context.Background())
	}

	db, err := client.CreateDatabase(context.Background(), dbName, nil)
	if err != nil {
		log.Fatalf("Failed to create database: %v\n", err)
	}
	return db
}

func evtKey(i int, j int) string {
	return fmt.Sprintf("%d,%d", i, j)
}

/*
returns db and graph
*/
func createGraph(client driver.Client, dbConsts DBConsts) (driver.Database, driver.Graph, driver.Graph) {
	db := getOrCreateDB(client, dbConsts.DB)

	txnDepEdgeDef := driver.EdgeDefinition{
		Collection: dbConsts.TxnDepEdge,
		From:       []string{dbConsts.TxnNode},
		To:         []string{dbConsts.TxnNode},
	}

	evtDepEdgeDef := driver.EdgeDefinition{
		Collection: dbConsts.EvtDepEdge,
		From:       []string{dbConsts.WriteEvtNode, dbConsts.ReadEvtNode},
		To:         []string{dbConsts.WriteEvtNode, dbConsts.ReadEvtNode},
	}

	_, err := db.CreateCollection(context.Background(), dbConsts.TxnDepEdge, &driver.CreateCollectionOptions{
		Type:           driver.CollectionTypeEdge,
		NumberOfShards: 1, // we put all txn nodes on the same shard
		ShardKeys:      []string{"_from"},
	})

	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}

	_, err = db.CreateCollection(context.Background(), dbConsts.EvtDepEdge, &driver.CreateCollectionOptions{
		Type:           driver.CollectionTypeEdge,
		NumberOfShards: 1,
		ShardKeys:      []string{"_from"},
	})

	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}

	_, err = db.CreateCollection(context.Background(), dbConsts.TxnNode, &driver.CreateCollectionOptions{
		NumberOfShards: 1,
	})

	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}

	_, err = db.CreateCollection(context.Background(), dbConsts.WriteEvtNode, &driver.CreateCollectionOptions{
		NumberOfShards: 1,
	})

	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}

	_, err = db.CreateCollection(context.Background(), dbConsts.ReadEvtNode, &driver.CreateCollectionOptions{
		NumberOfShards: 1,
	})

	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}

	txnGraphOpts := driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{
			txnDepEdgeDef,
		},
	}

	evtGraphOpts := driver.CreateGraphOptions{
		EdgeDefinitions: []driver.EdgeDefinition{
			evtDepEdgeDef,
		},
	}

	txnGraph, err := db.CreateGraphV2(context.Background(), dbConsts.TxnGraph, &txnGraphOpts)
	if err != nil {
		log.Fatalf("Failed to create graph: %v\n", err)
	}

	evtGraph, err := db.CreateGraphV2(context.Background(), dbConsts.EvtGraph, &evtGraphOpts)
	if err != nil {
		log.Fatalf("Failed to create graph: %v\n", err)
	}

	return db, txnGraph, evtGraph
}

/*
create nodes: txns, writeEvts & readEvts
*/
func createNodes(txnGraph driver.Graph, evtGraph driver.Graph, okHistory core.History, dbConsts DBConsts) []int {
	txns := make([]TxnNode, 0, len(okHistory))
	txnIds := make([]int, 0, len(okHistory))
	// init by assuming each txn has one write, two reads on avg
	writeEvts := make([]WriteEvt, 0, len(okHistory))
	readEvts := make([]ReadEvt, 0, len(okHistory)*2)
	for _, op := range okHistory {
		txnId := op.Index.MustGet()
		txnIds = append(txnIds, txnId)
		txns = append(txns, TxnNode{
			strconv.Itoa(txnId), // will panic if not found
		})

		writeIdxCounter := make(map[string]int)
		lastwriteMap := make(map[string]int)

		for j, v := range *op.Value {
			if v.IsRead() {
				readVal := v.GetValue()
				// we use zero-value as the default "start" value here
				if readVal == nil {
					readVal = 0
				}
				// mark those "first reads" with index 0
				readEvts = append(readEvts, ReadEvt{
					evtKey(txnId, j),
					v.GetKey(),
					readVal.(int),
				})
			} else if v.IsWrite() {
				writeEvts = append(writeEvts, WriteEvt{
					evtKey(txnId, j),
					v.GetKey(),
					v.GetValue().(int),
					writeIdxCounter[v.GetKey()],
				})
				writeIdxCounter[v.GetKey()]++
				lastwriteMap[v.GetKey()] = len(writeEvts) - 1
			}
		}
		// mark those "last writes" with index - 1
		for _, k := range lastwriteMap {
			writeEvts[k].Index = -1
		}
	}

	txnNodes, err := txnGraph.VertexCollection(context.Background(), dbConsts.TxnNode)
	if err != nil {
		log.Fatalf("Failed to get node collection: %v\n", err)
	}

	_, _, err = txnNodes.CreateDocuments(context.Background(), txns)
	if err != nil {
		log.Fatalf("Failed to create nodes: %v\n", err)
	}

	writeEvtNodes, err := evtGraph.VertexCollection(context.Background(), dbConsts.WriteEvtNode)
	if err != nil {
		log.Fatalf("Failed to get node collection: %v\n", err)
	}

	_, _, err = writeEvtNodes.CreateDocuments(context.Background(), writeEvts)
	if err != nil {
		log.Fatalf("Failed to create nodes: %v\n", err)
	}

	readEvtNodes, err := evtGraph.VertexCollection(context.Background(), dbConsts.ReadEvtNode)
	if err != nil {
		log.Fatalf("Failed to get node collection: %v\n", err)
	}

	_, _, err = readEvtNodes.CreateDocuments(context.Background(), readEvts)
	if err != nil {
		log.Fatalf("Failed to create nodes: %v\n", err)
	}

	return txnIds
}

type ReadEvtsInfo struct {
	Obj    string          `json:"obj"`
	Traces []ReadEvtsTrace `json:"traces"`
}

type ReadEvtsTrace struct {
	Val int      `json:"val"`
	Ids []string `json:"ids"`
}

/*
returns an array of read-events info
(with obj and traces as defined above)
*/
/*
FOR e1 IN r_evt
	COLLECT obj = e1.obj INTO objs
	RETURN { obj, traces: (
		FOR e2 in objs[*].e1
			COLLECT val = e2.v INTO vals
			RETURN { val, ids: vals[*].e2._id }
	)}
*/

func queryReadEvts(db driver.Database, dbConsts DBConsts) map[string]map[int][]string {
	query := fmt.Sprintf(`
		FOR e1 IN %s
			COLLECT obj = e1.obj INTO objs
			RETURN { obj, traces: (
				FOR e2 in objs[*].e1
					COLLECT val = e2.v INTO vals
					RETURN { val, ids: vals[*].e2._id }
			)}
	`, dbConsts.ReadEvtNode)

	cursor, err := db.Query(context.Background(), query, nil)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	defer cursor.Close()

	readMap := make(map[string]map[int][]string)

	for {
		var info ReadEvtsInfo
		_, err := cursor.ReadDocument(context.Background(), &info)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			obj := info.Obj
			if _, ok := readMap[obj]; !ok {
				readMap[obj] = make(map[int][]string)
			}
			for _, trace := range info.Traces {
				readMap[obj][trace.Val] = trace.Ids
			}
		}
	}
	return readMap
}

type WriteEvtsInfo struct {
	Obj  string             `json:"obj"`
	Evts []WriteEvtsElement `json:"evts"`
}

type WriteEvtsElement struct {
	Element  int      `json:"element"`
	Ids      []string `json:"ids"`
	WriteIdx []int    `json:"write_idx"`
}

/*
returns a write map {obj1: {key1: id1, key2: id2, ...}, ...}
*/
func queryWriteEvts(db driver.Database, dbConsts DBConsts) (map[string]map[int]string, map[string]map[int]bool) {
	query := fmt.Sprintf(`
		FOR e1 IN %s
			COLLECT obj = e1.obj into objs
			RETURN { obj, evts: (
				FOR e2 in objs[*].e1
					COLLECT element = e2.arg INTO elements
					RETURN { element, ids: elements[*].e2._id, write_idx: elements[*].e2.index }
			)}
	`, dbConsts.WriteEvtNode)

	cursor, err := db.Query(context.Background(), query, nil)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	defer cursor.Close()

	writeMap := make(map[string]map[int]string)
	// intermediate writes or not: with index != -1, intermediate writes
	itmdMap := make(map[string]map[int]bool)

	for {
		var info WriteEvtsInfo
		_, err := cursor.ReadDocument(context.Background(), &info)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			obj := info.Obj
			if _, ok := writeMap[obj]; !ok {
				writeMap[obj] = make(map[int]string)
				itmdMap[obj] = make(map[int]bool)
			}
			for _, evt := range info.Evts {
				if len(evt.Ids) == 1 {
					writeMap[obj][evt.Element] = evt.Ids[0]
					if evt.WriteIdx[0] != -1 {
						itmdMap[obj][evt.Element] = true
					}
				} else {
					log.Fatalf("Anomaly: Multiple events %v write the same value %v to the same object %v.\n",
						evt.Ids, evt.Element, obj)
				}
			}
		}
	}
	return writeMap, itmdMap
}

type EvtDepEdge struct {
	From string `json:"_from"`
	To   string `json:"_to"`
	Obj  string `json:"obj"`
	Type string `json:"type"`
}

type G1Anomalies struct {
	G1a bool
	G1b bool
}

// get txnId and EvtId
func parseEvtId(id string) (int, int, error) {
	idstrs := strings.Split(id, ",")
	txnIdStr, evtIdStr := strings.Split(idstrs[0], "/")[1], idstrs[1]
	txnId, err := strconv.Atoi(txnIdStr)
	if err != nil {
		return -1, -1, err
	}
	evtId, err := strconv.Atoi(evtIdStr)
	if err != nil {
		return -1, -1, err
	}
	return txnId, evtId, nil
}

// happens before within the same txn
func happensBefore(id1 string, id2 string) bool {
	txnId1, evtId1, err := parseEvtId(id1)
	if err != nil {
		return false
	}

	txnId2, evtId2, err := parseEvtId(id2)
	if err != nil {
		return false
	}

	return txnId1 == txnId2 && evtId1 < evtId2
}

func getEvtDepEdges(db driver.Database, wm WALWriteMap, dbConsts DBConsts) ([]EvtDepEdge, G1Anomalies) {
	readsInfoMap := queryReadEvts(db, dbConsts)
	writesInfoMap, itmdMap := queryWriteEvts(db, dbConsts)

	evtDepEdges := make([]EvtDepEdge, 0, len(readsInfoMap)*3)

	g1 := G1Anomalies{}

	// obj ~ G1a

	for obj, readSubMap := range readsInfoMap {
		if _, ok := wm[obj]; !ok && len(readSubMap) > 0 {
			for v, reads := range readSubMap {
				if v != 0 {
					g1.G1a = true
					log.Printf("G1a: The object %v does not have any writes (possibly aborted *internally*) but reads %v in events %v\n",
						obj, v, reads)
					break
				}
			}
		}
	}

	for obj, versions := range wm {
		readSubMap, writeSubMap := readsInfoMap[obj], writesInfoMap[obj]
		itmdSubMap := itmdMap[obj]

		// write :ok - writeSubMap must have and versions must have
		// write :fail - writeSubMap must not have and versions must not have
		// - raise G1a when writeSubMap doesn't have but versions has
		// - raise the non-recoverable anomaly when writeSubMap has but versions doesn't; exit immediately
		// - raise G1b if: cur w < cur r < next w, where < means happensBefore

		writesInLog := make(map[int]bool)

		/*
			deal with the first version, only wr edges
		*/
		if len(versions) == 0 {
			log.Fatalf("Anomaly: Broken WAL logs for object %v. Non-recoverable.\n", obj)
		}

		// ver, w, writeOk variables defined in advance
		var prevVer, ver, nextVer int
		var prevW, w, nextW string
		var prevWriteOk, writeOk, nextWriteOk bool

		ver = versions[0]
		w, writeOk = writeSubMap[ver]

		if len(versions) > 1 {
			// has next
			nextVer = versions[1]
			nextW, nextWriteOk = writeSubMap[nextVer]
		}
		writesInLog[ver] = true // make it a hashset for values in versions
		if writeOk {
			// rw: 0 -> cur w
			for _, prevR := range readSubMap[0] {
				if !happensBefore(prevR, w) {
					evtDepEdges = append(evtDepEdges, EvtDepEdge{prevR, w, obj, "rw"})
				}
			}

			// wr: cur w -> cur r's
			for _, r := range readSubMap[ver] {
				g1bRaised := false
				if nextWriteOk {
					// else branch is not handled as it will be handled in the next iteration
					// raise G1b if: cur w < cur r's < next w, where < means happensBefore
					allowed := happensBefore(w, r) && happensBefore(r, nextW)
					if itmd := itmdSubMap[ver]; itmd && !allowed {
						g1.G1b = true
						g1bRaised = true
						log.Printf("G1b: The object %v has an intermediate write %v in event %v and reads it in event %v\n",
							obj, ver, w, r)
					}
				}
				// ignore wr dependencies within the same txn
				if !g1bRaised && !happensBefore(w, r) {
					evtDepEdges = append(evtDepEdges, EvtDepEdge{w, r, obj, "wr"})
				}
			}
			prevVer, prevW, prevWriteOk = ver, w, writeOk
		} else {
			// raise G1a when writeSubMap doesn't have but versions has AND the value is read
			if len(readSubMap[ver]) > 0 {
				// aborted reads detected
				g1.G1a = true
				log.Printf("G1a: The object %v does not write %v (possibly aborted) but reads it in events %v\n",
					obj, ver, readSubMap[ver])
			}
		}

		// utilize the index of versions
		for i := 1; i < len(versions); i++ {
			// cur ver
			ver = versions[i]
			w, writeOk = writeSubMap[ver]
			// next ver
			if i+1 < len(versions) {
				// has next
				nextVer = versions[i+1]
				nextW, nextWriteOk = writeSubMap[nextVer]
			}
			writesInLog[ver] = true // make it a hashset for values in versions

			if writeOk {
				// rw: prev r's -> cur w
				if prevWriteOk {
					for _, prevR := range readSubMap[prevVer] {
						if !happensBefore(prevR, w) {
							evtDepEdges = append(evtDepEdges, EvtDepEdge{prevR, w, obj, "rw"})
						}
					}
				}

				// ww: prev w -> cur w
				if prevWriteOk && !happensBefore(prevW, w) {
					evtDepEdges = append(evtDepEdges, EvtDepEdge{prevW, w, obj, "ww"})
				}

				// wr: cur w -> cur r's
				for _, r := range readSubMap[ver] {
					g1bRaised := false
					if nextWriteOk {
						// else branch is not handled as it will be handled in the next iteration
						// raise G1b if: cur w < cur r's < next w, where < means happensBefore
						allowed := happensBefore(w, r) && happensBefore(r, nextW)
						if itmd := itmdSubMap[ver]; itmd && !allowed {
							g1.G1b = true
							g1bRaised = true
							log.Printf("G1b: The object %v has an intermediate write %v in event %v and reads it in event %v\n",
								obj, ver, w, r)
						}
					}
					// ignore wr dependencies within the same txn
					if !g1bRaised && !happensBefore(w, r) {
						evtDepEdges = append(evtDepEdges, EvtDepEdge{w, r, obj, "wr"})
					}
				}
				prevVer, prevW, prevWriteOk = ver, w, writeOk
			} else {
				// raise G1a when writeSubMap doesn't have but versions has AND the value is read
				if len(readSubMap[ver]) > 0 {
					// aborted reads detected
					g1.G1a = true
					log.Printf("G1a: The object %v does not write %v (possibly aborted) but reads it in events %v\n",
						obj, ver, readSubMap[ver])
				}
			}
		}

		// val ~ G1a
		for v := range readSubMap {
			// v == 0 is allowed
			if v != 0 && !writesInLog[v] {
				// raise G1a when readSubMap has but versions doesn't AND the value is read
				g1.G1a = true
				log.Printf("G1a: The object %v does not write %v (possibly aborted *internally*) but reads it in events %v\n",
					obj, v, readSubMap[v])
			}
		}

		for v := range writeSubMap {
			if !writesInLog[v] {
				// raise the non-recoverable anomaly when writeSubMap has but versions doesn't; exit immediately
				log.Printf("Warning: The write %v to object %v is not successful but the history records ok.\n",
					v, obj)
			}
		}
	}

	return evtDepEdges, g1
}

type TxnDepEdge struct {
	From    string `json:"_from"`
	To      string `json:"_to"`
	FromEvt string `json:"from_evt"`
	ToEvt   string `json:"to_evt"`
	Type    string `json:"type"`
}

func getTxnDepEdges(db driver.Database, evtDepEdges []EvtDepEdge, dbConst DBConsts) []TxnDepEdge {
	query := fmt.Sprintf(`
		LET projs = (
			FOR d IN %s
				LET from_txn = SPLIT(d._from, ["/", ","])[1]
				LET to_txn = SPLIT(d._to, ["/", ","])[1]
				FILTER from_txn != to_txn
				RETURN { _from: CONCAT("txn/", from_txn), _to: CONCAT("txn/", to_txn),
					from_evt: d._from, to_evt: d._to, type: d.type }
			)
	
		FOR proj IN projs
			COLLECT from = proj._from, to = proj._to, type = proj.type INTO groups = {
				"from_evt": proj.from_evt,
				"to_evt": proj.to_evt
			}
			RETURN {
				"_from": from,
				"_to": to,
				"type": type,
				"from_evt": groups[0].from_evt,
				"to_evt": groups[0].to_evt
			}
	`, dbConst.EvtDepEdge)

	cursor, err := db.Query(context.Background(), query, nil)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	defer cursor.Close()

	txnDepEdges := make([]TxnDepEdge, 0, len(evtDepEdges))
	txnDepEdgeId := 0

	for {
		var dep TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &dep)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			txnDepEdgeId++
			txnDepEdges = append(txnDepEdges, dep)
		}
	}

	return txnDepEdges
}

func addDepEdges(db driver.Database, txnGraph driver.Graph, evtGraph driver.Graph,
	dbConsts DBConsts, evtDepEdges []EvtDepEdge) {
	evtDepEdgeCol, _, err := evtGraph.EdgeCollection(context.Background(), dbConsts.EvtDepEdge)
	if err != nil {
		log.Fatalf("Failed to get edge collection: %v\n", err)
	}

	_, _, err = evtDepEdgeCol.CreateDocuments(context.Background(), evtDepEdges)
	if err != nil {
		log.Fatalf("Failed to create edges: %v\n", err)
	}

	// projections from evts to txns
	txnDepEdges := getTxnDepEdges(db, evtDepEdges, dbConsts)

	txnDepEdgeCol, _, err := txnGraph.EdgeCollection(context.Background(), dbConsts.TxnDepEdge)
	if err != nil {
		log.Fatalf("Failed to get edge collection: %v\n", err)
	}

	_, _, err = txnDepEdgeCol.CreateDocuments(context.Background(), txnDepEdges)
	if err != nil {
		log.Fatalf("Failed to create edges: %v\n", err)
	}
}
