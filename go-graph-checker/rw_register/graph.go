package rwregister

import (
	"context"
	"fmt"
	"log"
	"strconv"

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
	Key string `json:"_key"`
	Obj string `json:"obj"`
	Arg int    `json:"arg"`
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
		fmt.Println("db exists already and will be dropped first...")
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
func createNodes(txnGraph driver.Graph, evtGraph driver.Graph, okHistory core.History, dbConsts DBConsts, ignoreReads bool) []int {
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

		for j, v := range *op.Value {
			if v.IsRead() {
				if !ignoreReads {
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
				}
			} else if v.IsWrite() {
				writeEvts = append(writeEvts, WriteEvt{
					evtKey(txnId, j),
					v.GetKey(),
					v.GetValue().(int),
				})
			}
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
	Element int      `json:"element"`
	Ids     []string `json:"ids"`
}

/*
returns a write map {obj1: {key1: id1, key2: id2, ...}, ...}
*/
func queryWriteEvts(db driver.Database, dbConsts DBConsts) map[string]map[int]string {
	query := fmt.Sprintf(`
		FOR e1 IN %s
			COLLECT obj = e1.obj into objs
			RETURN { obj, evts: (
				FOR e2 in objs[*].e1
					COLLECT element = e2.arg INTO elements
					RETURN { element, ids: elements[*].e2._id }
			)}
	`, dbConsts.WriteEvtNode)

	cursor, err := db.Query(context.Background(), query, nil)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	defer cursor.Close()

	writeMap := make(map[string]map[int]string)

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
			}
			for _, evt := range info.Evts {
				if len(evt.Ids) == 1 {
					writeMap[obj][evt.Element] = evt.Ids[0]
				} else {
					log.Fatalf("Anomaly 1: Multiple events %v write %v to the same object %v\n",
						evt.Ids, evt.Element, obj)
				}
			}
		}
	}
	return writeMap
}

type EvtDepEdge struct {
	From string `json:"_from"`
	To   string `json:"_to"`
	Obj  string `json:"obj"`
	Type string `json:"type"`
}

func getEvtDepEdges(db driver.Database, wm WALWriteMap, dbConsts DBConsts) []EvtDepEdge {
	readsInfoMap := queryReadEvts(db, dbConsts)
	writesInfoMap := queryWriteEvts(db, dbConsts)

	evtDepEdges := make([]EvtDepEdge, 0, len(readsInfoMap)*3)

	for obj, versions := range wm {
		readSubMap, writeSubMap := readsInfoMap[obj], writesInfoMap[obj]
		prev := 0
		// rw, update, ww, wr, reads
		for _, cur := range versions {
			// rw: prev r's -> cur w
			w := writeSubMap[cur]
			for _, r := range readSubMap[prev] {
				evtDepEdges = append(evtDepEdges, EvtDepEdge{r, w, obj, "rw"})
			}
			// ww: prev w -> cur w
			if prev > 0 {
				evtDepEdges = append(evtDepEdges, EvtDepEdge{writeSubMap[prev], w, obj, "ww"})
			}
			// wr: cur w -> cur r's
			for _, r := range readSubMap[cur] {
				evtDepEdges = append(evtDepEdges, EvtDepEdge{w, r, obj, "wr"})
			}
			prev = cur
		}
	}

	return evtDepEdges
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
