package listappend

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/core"
	"github.com/jasonqiu98/anti-pattern-graph-checker-single/go-elle/txn"
)

func ConstructGraph(opts txn.Opts, history core.History, dbConsts DBConsts) (driver.Database, []int, G1Anomalies) {
	// collect ok histories
	history = preProcessHistory(history)
	okHistory := core.FilterOkHistory(history)

	// create db instances and graph
	client := startClient(dbConsts.Host, dbConsts.Port)
	db, txnGraph, evtGraph := createGraph(client, dbConsts)

	// create nodes
	txnIds := createNodes(txnGraph, evtGraph, okHistory, dbConsts)

	// create evt and txn dependency edges
	evtDepEdges, g1 := getEvtDepEdges(db, dbConsts)
	addDepEdges(db, txnGraph, evtGraph, dbConsts, evtDepEdges)

	return db, txnIds, g1
}

type ArangoPath struct {
	Edges    []TxnDepEdge `json:"edges"`
	Vertices []TxnNode    `json:"vertices"`
}

func cycleToStr(cycle []TxnDepEdge) string {
	if len(cycle) == 0 {
		log.Fatalf("Failed to convert cycle to string\n")
	}
	var pathBuilder strings.Builder
	pathBuilder.WriteString(fmt.Sprintf("T%s", strings.Split(cycle[0].From, "/")[1]))
	for _, e := range cycle {
		pathBuilder.WriteString(fmt.Sprintf(" (%s) T%s", e.Type, strings.Split(e.To, "/")[1]))
	}
	return pathBuilder.String()
}

const (
	MIN_DEPTH           = 2 // for sv, sv-filter, sv-random
	MAX_DEPTH_SV_SIMPLE = 4 // for sv, sv-random
	MAX_DEPTH_SV        = 4 // for sv-filter
)

func IsolationLevelChecker(db driver.Database, dbConsts DBConsts, txnIds []int, output bool, level string, mode string) (bool, []TxnDepEdge) {
	switch level {
	case "ser", "SER", "serializabilty", "SERIALIZABILITY":
		switch mode {
		case "sv":
			return CheckSERSV(db, dbConsts, txnIds, output)
		case "sv-filter":
			return CheckSERSVFilter(db, dbConsts, txnIds, output)
		case "sv-random":
			return CheckSERSVRandom(db, dbConsts, txnIds, output)
		case "sp", "sp-allcycles":
			return CheckSERSP(db, dbConsts, txnIds, output)
		default:
			log.Fatalf("invalid mode: %s, not from any of the following:\nsv, sv-filter, sv-random, sp, sp-allcycles\n", mode)
			return false, []TxnDepEdge{}
		}
	case "si", "SI", "snapshot isolation", "SNAPSHOT ISOLATION":
		switch mode {
		case "sv":
			return CheckSISV(db, dbConsts, txnIds, output)
		case "sv-filter":
			return CheckSISVFilter(db, dbConsts, txnIds, output)
		case "sv-random":
			return CheckSISVRandom(db, dbConsts, txnIds, output)
		case "sp":
			return CheckSISP(db, dbConsts, txnIds, output)
		case "sp-allcycles":
			return CheckSISPAllCycles(db, dbConsts, txnIds, output)
		default:
			log.Fatalf("invalid mode: %s, not from any of the following:\nsv, sv-filter, sv-random, sp, sp-allcycles\n", mode)
			return false, []TxnDepEdge{}
		}
	case "psi", "PSI", "parallel snapshot isolation", "PARALLEL SNAPSHOT ISOLATION":
		switch mode {
		case "sv":
			return CheckPSISV(db, dbConsts, txnIds, output)
		case "sv-filter":
			return CheckPSISVFilter(db, dbConsts, txnIds, output)
		case "sv-random":
			return CheckPSISVRandom(db, dbConsts, txnIds, output)
		case "sp":
			return CheckPSISP(db, dbConsts, txnIds, output)
		case "sp-allcycles":
			return CheckPSISPAllCycles(db, dbConsts, txnIds, output)
		default:
			log.Fatalf("invalid mode: %s, not from any of the following:\nsv, sv-filter, sv-random, sp, sp-allcycles\n", mode)
			return false, []TxnDepEdge{}
		}
	case "pl-2", "PL-2":
		switch mode {
		case "sv":
			return CheckPL2SV(db, dbConsts, txnIds, output)
		case "sv-filter":
			return CheckPL2SVFilter(db, dbConsts, txnIds, output)
		case "sv-random":
			return CheckPL2SVRandom(db, dbConsts, txnIds, output)
		case "sp":
			return CheckPL2SP(db, dbConsts, txnIds, output)
		case "sp-allcycles":
			return CheckPL2SPAllCycles(db, dbConsts, txnIds, output)
		default:
			log.Fatalf("invalid mode: %s, not from any of the following:\nsv, sv-filter, sv-random, sp, sp-allcycles\n", mode)
			return false, []TxnDepEdge{}
		}
	case "pl-1", "PL-1":
		switch mode {
		case "sv":
			return CheckPL1SV(db, dbConsts, txnIds, output)
		case "sv-filter":
			return CheckPL1SVFilter(db, dbConsts, txnIds, output)
		case "sv-random":
			return CheckPL1SVRandom(db, dbConsts, txnIds, output)
		case "sp":
			return CheckPL1SP(db, dbConsts, txnIds, output)
		case "sp-allcycles":
			return CheckPL1SPAllCycles(db, dbConsts, txnIds, output)
		default:
			log.Fatalf("invalid mode: %s, not from any of the following:\nsv, sv-filter, sv-random, sp, sp-allcycles\n", mode)
			return false, []TxnDepEdge{}
		}
	default:
		log.Fatalf("invalid level: %s, not from any of the following:\n[ser, SER, serializabilty, SERIALIZABILITY, si, SI, snapshot isolation, SNAPSHOT ISOLATION, psi, PSI, parallel snapshot isolation, PARALLEL SNAPSHOT ISOLATION, pl-2, PL-2, pl-1, PL-1]\n", mode)
		return false, []TxnDepEdge{}
	}
}

/*
-----------------------------------------------ANTI-PATTERNS-------------------------------------------------
*/

// any cycle violates SER / PL-3

// any cycle without at least two consecutive RW edges violates SI
// any cycle with at least two consecutive RW edges means it is NOT an anti-pattern
func isAntiPatternSI(cycle []TxnDepEdge) bool {
	for i, edge := range cycle {
		if edge.Type == "rw" && cycle[(i+1)%len(cycle)].Type == "rw" {
			return false
		}
	}
	return true
}

// any cycle without at least two RW edges violates PSI
func isAntiPatternPSI(cycle []TxnDepEdge) bool {
	counter := 0
	for _, edge := range cycle {
		if edge.Type == "rw" {
			counter++
			if counter == 2 {
				return false
			}
		}
	}
	return true
}

// G1c: any cycle without rw edges violates PL-2
// any cycle with any rw edge means it is NOT an anti-pattern
func isAntiPatternPL2(cycle []TxnDepEdge) bool {
	for _, edge := range cycle {
		if edge.Type == "rw" {
			return false
		}
	}
	return true
}

// G0: any cycle with only ww edges violates PL-1
// any cycle with egdes of types other than ww means it is NOT an anti-pattern
func isAntiPatternPL1(cycle []TxnDepEdge) bool {
	for _, edge := range cycle {
		if edge.Type != "ww" {
			return false
		}
	}
	return true
}

/*
-----------------------------------------------DETAILS OF CHECKERS-------------------------------------------------
*/

func CheckSERSV(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				FILTER edge._to == start._id
				LIMIT 1
				RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SER: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of SER detected by SV.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

/*
	We do not use the PRUNE keyword (https://www.arangodb.com/docs/3.9/aql/graphs-traversals.html#pruning)
	because pruning does not take effect on the return result. Although it may achieve the early stopping
	as it claims, usually we need to combine the usage of PRUNE and that of FILTER to achieve a correct
	result set, which may induce the same filtering twice. This is because PRUNE just stops in the middle
	of traversal only if a certain condition is found to be satisfied. Otherwise, PRUNE will still reach
	the end of the current iteration to ensure the completeness of traversal, and therefore does not
	overall achieve our goal to filter the correct result set during iterations.

	Alternatively, "filtering on path"
	(https://www.arangodb.com/docs/3.9/aql/graphs-traversals.html#filtering-on-the-path-vs-filtering-on-vertices-or-edges)
	works in a simliar way so that the traversal stops early when the filtering condition is satisfied.
	It tackles the drawback of PRUNE and does not induce repeating filtering. Therefore, we shall use
	"filtering on path", like the query shown above.
*/

func CheckSERSVFilter(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				FILTER LAST(path.edges[*]._to) == start._id
				LIMIT 1
				RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SER: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of SER detected by SV-Filter.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckSERSVRandom(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND @start
				GRAPH %s
				FILTER LAST(path.edges[*]._to) == @start
				LIMIT 1
				RETURN path.edges
		`, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	starts := txnIds
	// iterate randomly after shuffling the index array slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(txnIds), func(i, j int) { starts[i], starts[j] = starts[j], starts[i] })

	bindVars := make(map[string]interface{})
	for _, start := range starts {
		bindVars["start"] = fmt.Sprintf("txn/%d", start)
		cursor, err := db.Query(context.Background(), query, bindVars)
		if err != nil {
			log.Fatalf("Failed to check SER: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle []TxnDepEdge
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				if output {
					log.Println("Anti-Patterns of SER detected by SV-Random.")
					log.Println(cycleToStr(cycle))
				}
				// will early stop once a cycle is detected
				return false, cycle
			}
		}
	}

	return true, nil
}

// SP / SP-AllCycles for SER
func CheckSERSP(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR edge IN %v
			FOR p IN OUTBOUND K_SHORTEST_PATHS
				edge._to TO edge._from
				GRAPH %v
				LIMIT 1
				RETURN {edges: UNSHIFT(p.edges, edge), vertices: UNSHIFT(p.vertices, p.vertices[LENGTH(p.vertices) - 1])}
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SER: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle ArangoPath
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if len(cycle.Edges) > 0 {
				if output {
					log.Println("Anti-Patterns of SER detected by SP / SP-AllCycles.")
					log.Println(cycleToStr(cycle.Edges))
				}
				return false, cycle.Edges
			}
		}
	}

	return true, nil
}

/*
Pregel - will not output any cycle, just for the API uniformity

	FOR t IN txn
		COLLECT cycle = t.scc INTO cycles
		FILTER LENGTH(cycles) > 1
		RETURN cycles[*].t._id
*/
func CheckSERPregel(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
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
					log.Println("Pregel finished.")
				}

				if driver.IsNoMoreDocuments(err) {
					return true, nil
				} else if err != nil {
					log.Fatalf("Cannot read return values: %v\n", err)
				} else {
					if output {
						log.Println("Anti-Patterns of SER detected by Arango-Pregel.")
						log.Println(cycle)
					}
					return false, nil
				}
			}
		} else if job.State == driver.PregelJobStateCanceled {
			log.Fatalf("Pregel SCC algorithm was canceled: %v\n", err)
		}
	}
}

func CheckSISV(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path IN %d..%d
			OUTBOUND start._id
			GRAPH %s
			FILTER edge._to == start._id AND NOT REGEX_TEST(CONCAT_SEPARATOR(" ", path.edges[*].type), "(^rw.*rw$|rw rw)")
			LIMIT 1
			RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SI: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of SI detected by SV.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckSISVFilter(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path IN %d..%d
			OUTBOUND start._id
			GRAPH %s
			FILTER LAST(path.edges[*]._to) == start._id AND NOT REGEX_TEST(CONCAT_SEPARATOR(" ", path.edges[*].type), "(^rw.*rw$|rw rw)")
			LIMIT 1
			RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SI: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of SI detected by SV-Filter.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckSISVRandom(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND @start
				GRAPH %s
				FILTER LAST(path.edges[*]._to) == @start AND NOT REGEX_TEST(CONCAT_SEPARATOR(" ", path.edges[*].type), "(^rw.*rw$|rw rw)")
				LIMIT 1
				RETURN path.edges
		`, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	starts := txnIds
	// iterate randomly after shuffling the index array slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(txnIds), func(i, j int) { starts[i], starts[j] = starts[j], starts[i] })

	bindVars := make(map[string]interface{})
	for _, start := range starts {
		bindVars["start"] = fmt.Sprintf("txn/%d", start)
		cursor, err := db.Query(context.Background(), query, bindVars)
		if err != nil {
			log.Fatalf("Failed to check SI: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle []TxnDepEdge
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				if output {
					log.Println("Anti-Patterns of SI detected by SV-Random.")
					log.Println(cycleToStr(cycle))
				}
				// will early stop once a cycle is detected
				return false, cycle
			}
		}
	}

	return true, nil
}

/*
direct query a type of cycle and return in ArangoDB format
*/
func CheckSISP(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		LET cycles = (
			FOR edge IN %v
				FOR p IN OUTBOUND K_SHORTEST_PATHS
					edge._to TO edge._from
					GRAPH %v
					RETURN {edges: UNSHIFT(p.edges, edge), vertices: UNSHIFT(p.vertices, p.vertices[LENGTH(p.vertices) - 1])}
		)
		
		FOR cycle IN cycles
			FILTER NOT REGEX_TEST(CONCAT_SEPARATOR(" ", cycle.edges[*].type), "(^rw.*rw$|rw rw)")
			LIMIT 1
			RETURN cycle
		
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SI: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle ArangoPath
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if len(cycle.Edges) > 0 {
				if output {
					log.Println("Anti-Patterns of SI detected by SP.")
					log.Println(cycleToStr(cycle.Edges))
				}
				return false, cycle.Edges
			}
		}
	}

	return true, nil
}

/*
query using BFS-based shortest path + parsing the cycle results

	FOR edge IN dep
		FOR p IN OUTBOUND K_SHORTEST_PATHS
			edge._to TO edge._from
			GRAPH txn_g
			RETURN UNSHIFT(p.edges, edge)
*/
func CheckSISPAllCycles(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR edge IN %s
			FOR p IN OUTBOUND K_SHORTEST_PATHS
				edge._to TO edge._from
				GRAPH %s
				RETURN UNSHIFT(p.edges, edge)
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check SI: %v\n", err)
	}

	defer cursor.Close()

	cycles := [][]TxnDepEdge{}
	antiPatternFound := false

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			cycles = append(cycles, cycle)
			if len(cycles) > 0 && isAntiPatternSI(cycles[len(cycles)-1]) {
				// found one anti-pattern
				antiPatternFound = true
				break
			}
		}
	}

	// if only one anti-pattern, do the final check
	if antiPatternFound || (len(cycles) > 0 && isAntiPatternSI(cycles[len(cycles)-1])) {
		if output {
			log.Println("Anti-Patterns of SI detected by SP-AllCycles.")
			log.Println(cycleToStr(cycles[len(cycles)-1]))
		}
		return false, cycles[len(cycles)-1]
	}

	return true, nil
}

func CheckPSISV(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path IN %d..%d
			OUTBOUND start._id
			GRAPH %s
			FILTER edge._to == start._id AND LENGTH(FOR e IN path.edges FILTER e.type == "rw" RETURN e) < 2
			LIMIT 1
			RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PSI: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of PSI detected by SV.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckPSISVFilter(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path IN %d..%d
			OUTBOUND start._id
			GRAPH %s
			FILTER LAST(path.edges[*]._to) == start._id AND LENGTH(FOR e IN path.edges FILTER e.type == "rw" RETURN e) < 2
			LIMIT 1
			RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PSI: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of PSI detected by SV-Filter.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckPSISVRandom(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND @start
				GRAPH %s
				FILTER LAST(path.edges[*]._to) == @start AND LENGTH(FOR e IN path.edges FILTER e.type == "rw" RETURN e) < 2
				LIMIT 1
				RETURN path.edges
		`, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	starts := txnIds
	// iterate randomly after shuffling the index array slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(txnIds), func(i, j int) { starts[i], starts[j] = starts[j], starts[i] })

	bindVars := make(map[string]interface{})
	for _, start := range starts {
		bindVars["start"] = fmt.Sprintf("txn/%d", start)
		cursor, err := db.Query(context.Background(), query, bindVars)
		if err != nil {
			log.Fatalf("Failed to check PSI: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle []TxnDepEdge
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				if output {
					log.Println("Anti-Patterns of PSI detected by SV-Random.")
					log.Println(cycleToStr(cycle))
				}
				// will early stop once a cycle is detected
				return false, cycle
			}
		}
	}

	return true, nil
}

func CheckPSISP(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		LET cycles = (
			FOR edge IN %v
				FOR p IN OUTBOUND K_SHORTEST_PATHS
					edge._to TO edge._from
					GRAPH %v
					RETURN {edges: UNSHIFT(p.edges, edge), vertices: UNSHIFT(p.vertices, p.vertices[LENGTH(p.vertices) - 1])}
		)
		
		FOR cycle IN cycles
			FILTER LENGTH(FOR e IN cycle.edges FILTER e.type == "rw" RETURN e) < 2
			LIMIT 1
			RETURN cycle
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PSI: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle ArangoPath
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if len(cycle.Edges) > 0 {
				if output {
					log.Println("Anti-Patterns of PSI detected by SP.")
					log.Println(cycleToStr(cycle.Edges))
				}
				return false, cycle.Edges
			}
		}
	}

	return true, nil
}

func CheckPSISPAllCycles(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR edge IN %s
			FOR p IN OUTBOUND K_SHORTEST_PATHS
				edge._to TO edge._from
				GRAPH %s
				RETURN UNSHIFT(p.edges, edge)
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PSI: %v\n", err)
	}

	defer cursor.Close()

	cycles := [][]TxnDepEdge{}
	antiPatternFound := false

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			cycles = append(cycles, cycle)
			if len(cycles) > 0 && isAntiPatternPSI(cycles[len(cycles)-1]) {
				// found one anti-pattern
				antiPatternFound = true
				break
			}
		}
	}

	// if only one anti-pattern, do the final check
	if antiPatternFound || (len(cycles) > 0 && isAntiPatternPSI(cycles[len(cycles)-1])) {
		if output {
			log.Println("Anti-Patterns of PSI detected by SP-AllCycles.")
			log.Println(cycleToStr(cycles[len(cycles)-1]))
		}
		return false, cycles[len(cycles)-1]
	}

	return true, nil
}

/*
G2: Anti-dependency Cycles [cycles with at least one RW edge]
FOR start IN txn
   FOR vertex, edge, path
       IN 2..5
       OUTBOUND start._id
       GRAPH dep
       FILTER path.edges[*].type ANY == "rw" AND edge._to == start._id
       RETURN path.edges

G1c: Circular Information Flow [cycles with only WW or WR edges]
FOR start IN txn
   FOR vertex, edge, path
       IN 2..5
       OUTBOUND start._id
       GRAPH dep
       FILTER path.edges[*].type NONE == "rw" AND edge._to == start._id
       RETURN path.edges

G0: Write Cycles [cycles with only WW edges]
Requires a new graph with only WW cycles

FOR start IN txn
    FOR vertex, edge, path
        IN 2..5
        OUTBOUND start._id
        GRAPH dep
        RETURN path.edges
*/

/*
the anti-pattern of PL-2 is G1 (G1a, G1b, G1c)
only G1c will be checked as G1a and G1b are ensured not to happen during graph construction
*/
func CheckPL2SV(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				FILTER path.edges[*].type NONE == "rw" AND edge._to == start._id
				LIMIT 1
				RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-2: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of PL-2 detected by SV.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

/*
the anti-pattern of PL-2 is G1 (G1a, G1b, G1c)
only G1c will be checked as G1a and G1b are ensured not to happen during graph construction
*/
func CheckPL2SVFilter(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				FILTER path.edges[*].type NONE == "rw" AND LAST(path.edges[*]._to) == start._id
				LIMIT 1
				RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-2: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of PL-2 detected by SV-Filter.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckPL2SVRandom(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND @start
				GRAPH %s
				FILTER path.edges[*].type NONE == "rw" and LAST(path.edges[*]._to) == @start
				LIMIT 1
				RETURN path.edges
		`, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	starts := txnIds
	// iterate randomly after shuffling the index array slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(txnIds), func(i, j int) { starts[i], starts[j] = starts[j], starts[i] })

	bindVars := make(map[string]interface{})
	for _, start := range starts {
		bindVars["start"] = fmt.Sprintf("txn/%d", start)
		cursor, err := db.Query(context.Background(), query, bindVars)
		if err != nil {
			log.Fatalf("Failed to check PL-2: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle []TxnDepEdge
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				if output {
					log.Println("Anti-Patterns of PL-2 detected by SV-Random.")
					log.Println(cycleToStr(cycle))
				}
				// will early stop once a cycle is detected
				return false, cycle
			}
		}
	}

	return true, nil
}

func CheckPL2SP(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		LET cycles = (
			FOR edge IN %v
				FILTER edge != "rw"
				FOR p IN OUTBOUND K_SHORTEST_PATHS
					edge._to TO edge._from
					GRAPH %v
					RETURN {edges: UNSHIFT(p.edges, edge), vertices: UNSHIFT(p.vertices, p.vertices[LENGTH(p.vertices) - 1])}
		)
		
		FOR cycle IN cycles
			FILTER cycle.edges[*].type NONE == "rw"
			LIMIT 1
			RETURN cycle
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-2: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle ArangoPath
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if len(cycle.Edges) > 0 {
				if output {
					log.Println("Anti-Patterns of PL-2 detected by SP.")
					log.Println(cycleToStr(cycle.Edges))
				}
				return false, cycle.Edges
			}
		}
	}

	return true, nil
}

func CheckPL2SPAllCycles(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR edge IN %s
			FILTER edge != "rw"
			FOR p IN OUTBOUND K_SHORTEST_PATHS
				edge._to TO edge._from
				GRAPH %s
				RETURN UNSHIFT(p.edges, edge)
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-2: %v\n", err)
	}

	defer cursor.Close()

	cycles := [][]TxnDepEdge{}
	antiPatternFound := false

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			cycles = append(cycles, cycle)
			if len(cycles) > 0 && isAntiPatternPL2(cycles[len(cycles)-1]) {
				// found one anti-pattern
				antiPatternFound = true
				break
			}
		}
	}

	// if only one anti-pattern, do the final check
	if antiPatternFound || (len(cycles) > 0 && isAntiPatternPL2(cycles[len(cycles)-1])) {
		if output {
			log.Println("Anti-Patterns of PL-2 detected by SP-AllCycles.")
			log.Println(cycleToStr(cycles[len(cycles)-1]))
		}
		return false, cycles[len(cycles)-1]
	}

	return true, nil
}

/*
with a new graph consisting of only WW edges
any cycle would violate PL-1
*/
func CheckPL1SV(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				FILTER path.edges[*].type ALL == "ww" AND edge._to == start._id
				LIMIT 1
				RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV_SIMPLE, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-1: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of PL-1 detected by SV.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckPL1SVFilter(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR start IN %s
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND start._id
				GRAPH %s
				FILTER path.edges[*].type ALL == "ww" AND LAST(path.edges[*]._to) == start._id
				LIMIT 1
				RETURN path.edges
		`, dbConsts.TxnNode, MIN_DEPTH, MAX_DEPTH_SV, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-1: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if output {
				log.Println("Anti-Patterns of PL-1 detected by SV-Filter.")
				log.Println(cycleToStr(cycle))
			}
			return false, cycle
		}
	}

	return true, nil
}

func CheckPL1SVRandom(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	minStep := 2
	maxStep := 3
	query := fmt.Sprintf(`
			FOR vertex, edge, path
				IN %d..%d
				OUTBOUND @start
				GRAPH %s
				FILTER path.edges[*].type ALL == "ww" AND LAST(path.edges[*]._to) == @start
				LIMIT 1
				RETURN path.edges
		`, minStep, maxStep, dbConsts.TxnGraph)

	starts := txnIds
	// iterate randomly after shuffling the index array slice
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(txnIds), func(i, j int) { starts[i], starts[j] = starts[j], starts[i] })

	bindVars := make(map[string]interface{})
	for _, start := range starts {
		bindVars["start"] = fmt.Sprintf("txn/%d", start)
		cursor, err := db.Query(context.Background(), query, bindVars)
		if err != nil {
			log.Fatalf("Failed to check PL-1: %v\n", err)
		}

		defer cursor.Close()

		for {
			var cycle []TxnDepEdge
			_, err := cursor.ReadDocument(context.Background(), &cycle)

			if driver.IsNoMoreDocuments(err) {
				break
			} else if err != nil {
				log.Fatalf("Cannot read return values: %v\n", err)
			} else {
				if output {
					log.Println("Anti-Patterns of PL-1 detected by SV-Random.")
					log.Println(cycleToStr(cycle))
				}
				// will early stop once a cycle is detected
				return false, cycle
			}
		}
	}

	return true, nil
}

func CheckPL1SPAllCycles(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		FOR edge IN %s
			FILTER edge.type == "ww"
			FOR p IN OUTBOUND K_SHORTEST_PATHS
				edge._to TO edge._from
				GRAPH %s
				RETURN UNSHIFT(p.edges, edge)
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-1: %v\n", err)
	}

	defer cursor.Close()

	cycles := [][]TxnDepEdge{}
	antiPatternFound := false

	for {
		var cycle []TxnDepEdge
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			cycles = append(cycles, cycle)
			if len(cycles) > 0 && isAntiPatternPL1(cycles[len(cycles)-1]) {
				// found one anti-pattern
				antiPatternFound = true
				break
			}
		}
	}

	// if only one anti-pattern, do the final check
	if antiPatternFound || (len(cycles) > 0 && isAntiPatternPL1(cycles[len(cycles)-1])) {
		if output {
			log.Println("Anti-Patterns of PL-1 detected by SP-AllCycles.")
			log.Println(cycleToStr(cycles[len(cycles)-1]))
		}
		return false, cycles[len(cycles)-1]
	}

	return true, nil
}

/*
direct query a type of cycle and return in ArangoDB format
*/
func CheckPL1SP(db driver.Database, dbConsts DBConsts, txnIds []int, output bool) (bool, []TxnDepEdge) {
	query := fmt.Sprintf(`
		LET cycles = (
			FOR edge IN %v
				FILTER edge.type == "ww"
				FOR p IN OUTBOUND K_SHORTEST_PATHS
					edge._to TO edge._from
					GRAPH %v
					RETURN {edges: UNSHIFT(p.edges, edge), vertices: UNSHIFT(p.vertices, p.vertices[LENGTH(p.vertices) - 1])}
		)
		
		FOR cycle IN cycles
			FILTER cycle.edges[*].type ALL == "ww"
			LIMIT 1
			RETURN cycle
		`, dbConsts.TxnDepEdge, dbConsts.TxnGraph)

	cursor, err := db.Query(context.Background(), query, nil)
	if err != nil {
		log.Fatalf("Failed to check PL-1: %v\n", err)
	}

	defer cursor.Close()

	for {
		var cycle ArangoPath
		_, err := cursor.ReadDocument(context.Background(), &cycle)

		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Fatalf("Cannot read return values: %v\n", err)
		} else {
			if len(cycle.Edges) > 0 {
				if output {
					log.Println("Anti-Patterns of PL-1 detected by SP.")
					log.Println(cycleToStr(cycle.Edges))
				}
				return false, cycle.Edges
			}
		}
	}

	return true, nil
}
