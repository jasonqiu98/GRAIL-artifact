package listappend

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	driver "github.com/arangodb/go-driver"
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

type AppendEvt struct {
	Key   string `json:"_key"`
	Obj   string `json:"obj"`
	Arg   int    `json:"arg"`
	Index int    `json:"index"` // -1 means last append
}

type ReadEvt struct {
	Key string `json:"_key"`
	Obj string `json:"obj"`
	V   []int  `json:"v"`
}

type DBConsts struct {
	Host          string
	Port          int
	DB            string
	TxnGraph      string
	EvtGraph      string
	TxnNode       string
	AppendEvtNode string
	ReadEvtNode   string
	TxnDepEdge    string
	EvtDepEdge    string
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
		From:       []string{dbConsts.AppendEvtNode, dbConsts.ReadEvtNode},
		To:         []string{dbConsts.AppendEvtNode, dbConsts.ReadEvtNode},
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

	_, err = db.CreateCollection(context.Background(), dbConsts.AppendEvtNode, &driver.CreateCollectionOptions{
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
create nodes: txns, appendEvts & readEvts
*/
func createNodes(txnGraph driver.Graph, evtGraph driver.Graph, okHistory core.History, dbConsts DBConsts) []int {
	txns := make([]TxnNode, 0, len(okHistory))
	txnIds := make([]int, 0, len(okHistory))
	// init by assuming each txn has one append, two reads on avg
	appendEvts := make([]AppendEvt, 0, len(okHistory))
	readEvts := make([]ReadEvt, 0, len(okHistory)*2)
	for _, op := range okHistory {
		txnId := op.Index.MustGet()
		txnIds = append(txnIds, txnId)
		txns = append(txns, TxnNode{
			strconv.Itoa(txnId), // will panic if not found
		})

		appendIdxCounter := make(map[string]int)
		lastAppendMap := make(map[string]int)

		for j, v := range *op.Value {
			if v.IsRead() {
				readVal := v.GetValue()
				if readVal == nil {
					readVal = make([]int, 0)
				}
				// mark those "first reads" with index 0
				readEvts = append(readEvts, ReadEvt{
					evtKey(txnId, j),
					v.GetKey(),
					readVal.([]int),
				})
			} else if v.IsAppend() {
				appendEvts = append(appendEvts, AppendEvt{
					evtKey(txnId, j),
					v.GetKey(),
					v.GetValue().(int),
					appendIdxCounter[v.GetKey()],
				})
				appendIdxCounter[v.GetKey()]++
				lastAppendMap[v.GetKey()] = len(appendEvts) - 1
			}
		}

		// mark those "last appends" with index - 1
		for _, k := range lastAppendMap {
			appendEvts[k].Index = -1
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

	appendEvtNodes, err := evtGraph.VertexCollection(context.Background(), dbConsts.AppendEvtNode)
	if err != nil {
		log.Fatalf("Failed to get node collection: %v\n", err)
	}

	_, _, err = appendEvtNodes.CreateDocuments(context.Background(), appendEvts)
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

// types of query results

type ReadEvtsInfo struct {
	Obj    string          `json:"obj"`
	Traces []ReadEvtsTrace `json:"traces"`
}

type ReadEvtsTrace struct {
	Val []int    `json:"val"`
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
			SORT LENGTH(val) DESC
			RETURN { val, ids: vals[*].e2._id }
	)}
*/
func queryReadEvts(db driver.Database, dbConsts DBConsts) (arr []ReadEvtsInfo) {
	query := fmt.Sprintf(`
		FOR e1 IN %s
			COLLECT obj = e1.obj INTO objs
			RETURN { obj, traces: (
				FOR e2 in objs[*].e1
					COLLECT val = e2.v INTO vals
					SORT LENGTH(val) DESC
					RETURN { val, ids: vals[*].e2._id }
			)}
	`, dbConsts.ReadEvtNode)

	cursor, err := db.Query(context.Background(), query, nil)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	defer cursor.Close()

	for {
		var info ReadEvtsInfo
		_, err := cursor.ReadDocument(context.Background(), &info)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			arr = append(arr, info)
		}
	}
	return
}

type AppendEvtsInfo struct {
	Obj  string              `json:"obj"`
	Evts []AppendEvtsElement `json:"evts"`
}

type AppendEvtsElement struct {
	Element   int      `json:"element"`
	Ids       []string `json:"ids"`
	AppendIdx []int    `json:"append_idx"`
}

/*
returns an append map {obj1: {key1: id1, key2: id2, ...}, ...}
*/
func queryAppendEvts(db driver.Database, dbConsts DBConsts) (map[string]map[int]string, map[string]map[int]bool) {
	query := fmt.Sprintf(`
		FOR e1 IN %s
			COLLECT obj = e1.obj into objs
			RETURN { obj, evts: (
				FOR e2 in objs[*].e1
					COLLECT element = e2.arg INTO elements
					RETURN { element, ids: elements[*].e2._id, append_idx: elements[*].e2.index }
			)}
	`, dbConsts.AppendEvtNode)

	cursor, err := db.Query(context.Background(), query, nil)

	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}

	defer cursor.Close()

	appendMap := make(map[string]map[int]string)
	// intermediate appends or not: with index != -1, intermediate appends
	itmdMap := make(map[string]map[int]bool)

	for {
		var info AppendEvtsInfo
		_, err := cursor.ReadDocument(context.Background(), &info)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			obj := info.Obj
			if _, ok := appendMap[obj]; !ok {
				appendMap[obj] = make(map[int]string)
				itmdMap[obj] = make(map[int]bool)
			}
			for _, evt := range info.Evts {
				if len(evt.Ids) == 1 {
					appendMap[obj][evt.Element] = evt.Ids[0]
					if evt.AppendIdx[0] != -1 {
						itmdMap[obj][evt.Element] = true
					}
				} else {
					log.Fatalf("Anomaly: Multiple events %v append the same value %v to the same object %v. Non-recoverable.\n",
						evt.Ids, evt.Element, obj)
				}
			}
		}
	}
	return appendMap, itmdMap
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

func getEvtDepEdges(db driver.Database, dbConsts DBConsts) ([]EvtDepEdge, G1Anomalies) {
	readEvtsInfoArr := queryReadEvts(db, dbConsts)
	appendMap, itmdMap := queryAppendEvts(db, dbConsts)

	evtDepEdges := make([]EvtDepEdge, 0, len(readEvtsInfoArr)*3)
	evtDepEdgeId := 0

	g1 := G1Anomalies{}

	// iterate over the array of read-events info.
	for _, info := range readEvtsInfoArr {
		obj := info.Obj
		objAppendMap, ok := appendMap[obj]
		traces := info.Traces
		if len(traces) == 0 {
			// no traces found, so no edges could be interpreted
			continue
		}

		if !ok {
			// the objects are not written by any known events
			// the read values must be the initial value (i.e. an empty array)
			if len(traces) == 1 && len(traces[0].Val) == 0 {
				// a valid trace, so we just ignore
				continue
			} else {
				for _, trace := range traces {
					if len(trace.Val) > 0 {
						g1.G1a = true
						log.Printf("G1a.1: The object %v has no append events (possibly aborted), but reads %v in events %v\n",
							obj, trace.Val, trace.Ids)
					}
				}
				// for some reason this info has an empty trace but still valid
				// normally the code won't reach here
				continue
			}
		}

		// must get, as both maps were created at the same time
		objItmdMap := itmdMap[obj]

		// the first trace
		longerVal := traces[0].Val
		longerRidArr := traces[0].Ids
		if len(longerVal) == 0 {
			// no new appends are read
			// however, there may exist further appends
			// add rw edges
			for _, laterAid := range objAppendMap {
				// rw
				// we need to construct rw dependencies
				// between longerAppended and the values appended later
				for _, rid := range longerRidArr {
					if !happensBefore(rid, laterAid) {
						evtDepEdges = append(evtDepEdges, EvtDepEdge{
							rid,
							laterAid,
							obj,
							"rw",
						})
						evtDepEdgeId++
					}
				}
			}
			// skip further steps
			continue
		}
		longerAppended := longerVal[len(longerVal)-1]
		longerAid, longerAidOk := objAppendMap[longerAppended]

		if !longerAidOk {
			g1.G1a = true
			log.Printf("G1a.2: The object %v has no %v in its append events (possibly aborted), but reads %v in events %v\n",
				obj, longerAppended, longerVal, longerRidArr)
		}

		valueSet := make(map[int]interface{})
		for _, v := range longerVal {
			valueSet[v] = nil
		}
		laterAidMap := make(map[string]bool) // rid has a laterAid
		for laterAppended, laterAid := range objAppendMap {
			// rw
			// we need to construct rw dependencies
			// between longerAppended and the values appended later
			if _, ok := valueSet[laterAppended]; !ok {
				for _, rid := range longerRidArr {
					if !happensBefore(rid, laterAid) {
						evtDepEdges = append(evtDepEdges, EvtDepEdge{
							rid,
							laterAid,
							obj,
							"rw",
						})
						evtDepEdgeId++
					} else {
						laterAidMap[rid] = true
					}
				}
				// ww
				// between longerAppended and the values appended later
				if longerAidOk && !happensBefore(longerAid, laterAid) {
					evtDepEdges = append(evtDepEdges, EvtDepEdge{
						longerAid,
						laterAid,
						obj,
						"ww",
					})
					evtDepEdgeId++
				}
			}
		}
		// wr
		// only when longerAid is ok

		if longerAidOk {
			for _, rid := range longerRidArr {
				// only longerAid < rid < laterAid is allowed
				allowed := happensBefore(longerAid, rid) && laterAidMap[rid]
				if longerItmd := objItmdMap[longerAppended]; longerItmd && !allowed {
					g1.G1b = true
					log.Printf("G1b.1: The object %v has an intermediate append %v in event %v and reads %v in event %v\n",
						obj, longerAppended, longerAid, longerVal, rid)
				} else {
					if !happensBefore(longerAid, rid) {
						evtDepEdges = append(evtDepEdges, EvtDepEdge{
							longerAid,
							rid,
							obj,
							"wr",
						})
						evtDepEdgeId++
					}
				}

			}
		}

		for i := 1; i < len(traces); i++ {
			trace := traces[i]
			val := trace.Val
			ridArr := trace.Ids

			// e.g. we have val: (i_1, i_2, ..., i_m)
			// and longerVal: (i_m+1, i_m+2, ..., i_n)

			if isPrefix(val, longerVal) {
				// rw
				// nextAid: the exact next append after val, i.e., i_m ->(rw) i_m+1
				nextAppended := longerVal[len(val)]
				nextAid, nextAidOk := objAppendMap[nextAppended]
				if !nextAidOk {
					g1.G1a = true
					log.Printf("G1a.3: The object %v has no %v in its append events (possibly aborted), but reads %v in events %v\n",
						obj, nextAppended, longerVal, longerRidArr)
				} else {
					for _, rid := range ridArr {
						if !happensBefore(rid, nextAid) {
							evtDepEdges = append(evtDepEdges, EvtDepEdge{
								rid,
								nextAid,
								obj,
								"rw",
							})
							evtDepEdgeId++
						}
					}
				}

				if len(val) > 0 {
					// ww
					// i_j -> i_j+1, j = m, m+1, .., n-1
					// nextAid updated from n to m+1
					nextAid, nextAidOk := longerAid, longerAidOk
					for j := len(longerVal) - 2; j >= len(val)-1; j-- {
						appended := longerVal[j]
						aid, aidOk := objAppendMap[appended]
						if !aidOk {
							g1.G1a = true
							log.Printf("G1a.4: The object %v has no %v in its append events (possibly aborted), but reads %v in events %v\n",
								obj, appended, longerVal, longerRidArr)
						} else {
							// aid ok, but nextAid might not be okay
							if nextAidOk && !happensBefore(aid, nextAid) {
								evtDepEdges = append(evtDepEdges, EvtDepEdge{
									aid,
									nextAid,
									obj,
									"ww",
								})
								evtDepEdgeId++
							}
						}

						nextAid, nextAidOk = aid, aidOk
					}

					appended := val[len(val)-1]
					// might not successfully read
					// but G1a check has already been done in the ww step
					aid, aidOk := objAppendMap[appended]
					nextAid, nextAidOk = objAppendMap[longerVal[len(val)]]

					if aidOk && nextAidOk {
						// wr
						for _, rid := range ridArr {
							// only aid < rid < nextAid is allowed
							allowed := happensBefore(aid, rid) && happensBefore(rid, nextAid)
							if itmd := objItmdMap[appended]; itmd && !allowed {
								g1.G1b = true
								log.Printf("G1b.2: The object %v has an intermediate append %v in event %v and reads %v in event %v\n",
									obj, appended, aid, val, rid)
							} else {
								if !happensBefore(aid, rid) {
									evtDepEdges = append(evtDepEdges, EvtDepEdge{
										aid,
										rid,
										obj,
										"wr",
									})
									evtDepEdgeId++
								}
							}
						}
					}

					// update pointers
					longerVal = val
					longerRidArr = ridArr
					longerAid, longerAidOk = aid, aidOk
				} else {
					// reaches the empty array
					// the break clause can also be neglected

					break
				}

			} else {
				// once a value is appended, it cannot be removed
				// this case violates the non-traceable property
				log.Fatalf("Anomaly 2: %v read by events %v is not a prefix of %v read by events %v (inconsistent read events under object %v). Non-traceable.\n",
					val, ridArr, longerVal, longerRidArr, obj)
			}
		}

		// add missing "ww" edges
		for i := len(longerVal) - 2; i >= 0; i-- {
			appended := longerVal[i]
			aid, aidOk := objAppendMap[appended]
			if !aidOk {
				g1.G1a = true
				log.Printf("G1a.5: The object %v has no %v in its append events (possibly aborted), but reads %v in events %v\n",
					obj, appended, longerVal, longerRidArr)
			} else {
				if longerAidOk && !happensBefore(aid, longerAid) {
					evtDepEdges = append(evtDepEdges, EvtDepEdge{
						aid,
						longerAid,
						obj,
						"ww",
					})
					evtDepEdgeId++
				}
			}

			longerAid, longerAidOk = aid, aidOk
		}
	}

	return evtDepEdges, g1
}

func isPrefix(v1 []int, v2 []int) bool {
	if len(v1) >= len(v2) {
		return false
	}

	for i := 0; i < len(v1); i++ {
		if v1[i] != v2[i] {
			return false
		}
	}

	return true
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
