package rwregister

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/grail/anti-pattern-graph-checker-single/go-elle/core"
)

type record struct {
	name      string
	label     string
	height    float32
	color     string
	fontColor string
}

func (r record) String() string {
	return fmt.Sprintf(`%s [height=%.2f,shape=record,label="%s",color="%s",fontcolor="%s"]`, r.name, r.height, r.label, r.color, r.fontColor)
}

type edge struct {
	from      string
	to        string
	label     string
	color     string
	fontColor string
}

func (e edge) String() string {
	return fmt.Sprintf(`%s -> %s [label="%s",fontcolor="%s",color="%s"]`, e.from, e.to, e.label, e.color, e.fontColor)
}

var typeColor = map[core.OpType]string{
	core.OpTypeOk:   "#0058AD",
	core.OpTypeInfo: "#AC6E00",
	core.OpTypeFail: "#A50053",
}

func relColor(rel string) string {
	switch rel {
	case "ww":
		return "#C02700"
	case "wr":
		return "#C000A5"
	case "rw":
		return "#5B00C0"
	default:
		return "#585858"
	}

}

func getEdgeEnds(edge TxnDepEdge) (string, string) {
	return strings.Split(edge.From, "/")[1], strings.Split(edge.To, "/")[1]
}

func getEdgeMopIndices(edge TxnDepEdge) (string, string) {
	return strings.Split(edge.FromEvt, ",")[1], strings.Split(edge.ToEvt, ",")[1]
}

func renderOp(nodeMap map[core.Op]string, op core.Op) record {
	var labels []string
	for idx, mop := range *op.Value {
		labels = append(labels, fmt.Sprintf("<f%d> %s", idx, mop.String()))
	}
	return record{
		name:      nodeMap[op],
		label:     strings.Join(labels, "|"),
		height:    0.4,
		color:     typeColor[op.Type],
		fontColor: typeColor[op.Type],
	}
}

func renderEdge(nodeMap map[core.Op]string, e TxnDepEdge) edge {
	a, b := getEdgeEnds(e)
	ami, bmi := getEdgeMopIndices(e)
	an := fmt.Sprintf("T%s:f%s", a, ami)
	bn := fmt.Sprintf("T%s:f%s", b, bmi)
	return edge{
		from:      an,
		to:        bn,
		label:     fmt.Sprintf("T%s (%s) T%s", a, e.Type, b),
		color:     relColor(e.Type),
		fontColor: relColor(e.Type),
	}
}

func renderCycle(history core.History, cycle []TxnDepEdge, output bool) (string, error) {
	opMap := make(map[int]core.Op)
	for _, op := range history {
		opMap[op.Index.MustGet()] = op
	}

	tpl := []string{"digraph g {"}
	nodeMap := make(map[core.Op]string)

	var nodes []record
	var edges []edge

	for _, e := range cycle {
		// "txn/1" -> 1
		srcIdx, _ := getEdgeEnds(e)
		opIdx, err := strconv.Atoi(srcIdx)
		if err != nil {
			return "", err
		}
		op := opMap[opIdx]
		// assume the index is always present
		nodeMap[op] = fmt.Sprintf("T%d", op.Index.MustGet())
		nodes = append(nodes, renderOp(nodeMap, op))
	}

	for _, e := range cycle {
		edges = append(edges, renderEdge(nodeMap, e))
	}

	for _, node := range nodes {
		tpl = append(tpl, fmt.Sprintf("    %s", node.String()))
	}

	tpl = append(tpl, "\n")

	for _, e := range edges {
		tpl = append(tpl, fmt.Sprintf("    %s", e.String()))
	}

	tpl = append(tpl, "}")

	if output {
		fmt.Println("-----------------------------------")
		fmt.Println("graphviz code:")
		fmt.Println(strings.Join(tpl, "\n"))
		fmt.Println("-----------------------------------")
	}

	return strings.Join(tpl, "\n"), nil
}

func PlotCycle(history core.History, cycle []TxnDepEdge, directory string, filename string, output bool) error {
	g := graphviz.New()
	tpl, err := renderCycle(history, cycle, output)
	if err != nil {
		return err
	}
	graph, err := graphviz.ParseBytes([]byte(tpl))
	if err != nil {
		return err
	}
	svgPath := fmt.Sprintf("%s/%s.svg", directory, filename)
	if err := g.RenderFilename(graph, graphviz.SVG, svgPath); err != nil {
		return err
	}

	return nil
}
