package listappend

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func ConstructGraph(opts txn.Opts, history core.History, dbConsts DBConsts) (driver.Database, map[string]int) {
	// collect ok histories
	history = preProcessHistory(history)
	okHistory := core.FilterOkHistory(history)

	// create db instances and graph
	client := startClient(dbConsts.Host, dbConsts.Port)
	db, txnGraph, evtGraph := createGraph(client, dbConsts)

	// create nodes
	metadata := createNodes(txnGraph, evtGraph, okHistory, dbConsts)

	// create evt and txn dependency edges
	evtDepEdges := getEvtDepEdges(db, dbConsts)
	addDepEdges(db, txnGraph, evtGraph, dbConsts, evtDepEdges)

	return db, metadata
}

func CheckSERV1(db driver.Database, dbConsts DBConsts, metadata map[string]int, output bool) bool {
	minStep := 2
	maxStep := 5
	/* e.g.
	FOR start IN txn
		FOR vertex, edge, path
			IN 2..5
			OUTBOUND start._id
			GRAPH txn_g
			PRUNE cond = edge._to == start._id
			FILTER cond
			RETURN CONCAT_SEPARATOR("->", path.vertices[*]._key)
	*/

	/*
		FOR start IN txn
			FOR vertex, edge, path
				IN 2..5
				OUTBOUND start._id
				GRAPH txn_g
				PRUNE cond = POP(path.vertices[*]._id) ANY == vertex._id
				FILTER cond
				RETURN CONCAT_SEPARATOR("->", path.vertices[*]._key)
	*/
	query := fmt.Sprintf(`
		FOR start IN txn
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				PRUNE cond = edge._to == start._id
				FILTER cond
				RETURN CONCAT_SEPARATOR("->", path.vertices[*]._key)
		`, minStep, maxStep, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SER: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle string
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				fmt.Println("Cycle Detected by V1.")
				fmt.Println(cycle)
			}
			return false
		}
	}

	return true
}

func CheckSERV2(db driver.Database, dbConsts DBConsts, metadata map[string]int, output bool) bool {
	minStep := 2
	maxStep := 5
	/* e.g.
	FOR start IN txn
		FOR vertex, edge, path
			IN 2..5
			OUTBOUND start._id
			GRAPH txn_g
			PRUNE cond = edge._to == start._id
			FILTER cond
			RETURN CONCAT_SEPARATOR("->", path.vertices[*]._key)
	*/
	query := fmt.Sprintf(`
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND @start
				GRAPH %s
				PRUNE cond = edge._to == @start
				FILTER cond
				RETURN CONCAT_SEPARATOR("->", path.vertices[*]._key)
		`, minStep, maxStep, dbConsts.TxnGraph)

	totalTxns := metadata["txns"]
	starts := make([]int, totalTxns)
	for i := 0; i < totalTxns; i++ {
		starts[i] = i
	}
	// iterate randomly after shuffling the index array slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(totalTxns, func(i, j int) { starts[i], starts[j] = starts[j], starts[i] })

	bindVars := make(map[string]interface{})
	for _, start := range starts {
		bindVars["start"] = fmt.Sprintf("txn/%d", start)
		cursor, err := db.Query(context.Background(), query, bindVars)
		if err != nil {
			log.Fatalf("Failed to check SER: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle string
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				if output {
					fmt.Println("Cycle Detected by V2.")
					fmt.Println(cycle)
				}
				// will early stop once a cycle is detected
				return false
			}
		}
	}

	return true
}

/*
Pregel
*/
func CheckSERPregel(db driver.Database, dbConsts DBConsts, metadata map[string]int, output bool) bool {
	jobId, err := db.StartJob(context.Background(), driver.PregelJobOptions{
		Algorithm: driver.PregelAlgorithmStronglyConnectedComponents,
		GraphName: dbConsts.TxnGraph,
		Params: map[string]interface{}{
			"resultField":       "scc",
			"shardKeyAttribute": "_from",
			"store":             true,
		},
	})

	if err != nil {
		log.Fatalf("Failed to start Pregel SCC algorithm: %v\n", err)
	}

	if len(jobId) == 0 {
		log.Fatalf("JobId is empty\n")
	}

	for {
		job, err := db.GetJob(context.Background(), jobId)

		if err != nil {
			log.Fatalf("Failed to get job: %v\n", err)
		}
		if jobId != job.ID {
			log.Fatalf("JobId mismatch\n")
		}
		if job.Reports == nil {
			log.Fatalf("Reports are empty\n")
		}

		if job.State == driver.PregelJobStateDone {
			// TODO: write a query to select all SCCs
			query := fmt.Sprintf(`
				FOR t IN %s
					COLLECT cycle = t.scc INTO cycles
					FILTER LENGTH(cycles) > 1
					RETURN cycle
			`, dbConsts.TxnNode)

			cursor, err := db.Query(context.Background(), query, nil)
			if err != nil {
				log.Fatalf("Failed to check SER: %v\n", err)
			}

			defer cursor.Close()

			for {
				var cycle int
				_, err := cursor.ReadDocument(context.Background(), &cycle)

				if output {
					fmt.Println("Pregel finished.")
				}

				if driver.IsNoMoreDocuments(err) {
					return true
				} else if err != nil {
					log.Fatalf("Cannot read return values: %v\n", err)
				} else {
					if output {
						fmt.Println("Cycle Detected by Pregel.")
						fmt.Println(cycle)
					}
					return false
				}
			}
		} else if job.State == driver.PregelJobStateCanceled {
			log.Fatalf("Pregel SCC algorithm was canceled: %v\n", err)
		}
	}
}
