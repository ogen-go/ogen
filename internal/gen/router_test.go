package gen

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ogen-go/ogen/internal/ir"
)

type GeneratedRouter struct {
}

type MatchResult struct {
	NotFound bool
	Node     int
	Args     []MatchArgument
}

type MatchArgument struct {
	Key   string
	Value []byte
}

func (r GeneratedRouter) Match(path []byte, args []MatchArgument) MatchResult {
	if len(path) == 0 {
		return MatchResult{NotFound: true}
	}
	// fmt.Printf("Match(%s)\n", path)
	var depth, parent, offset int
	var head []byte
Route:
	for {
		depth++
		if len(path) < offset+1 {
			return MatchResult{NotFound: true}
		}
		path = path[offset:]
		idx := bytes.IndexByte(path[offset+1:], '/')
		if idx < 0 {
			idx = len(path)
			offset = idx
		} else {
			offset += idx + 1
		}
		head = path[:offset]
		// fmt.Printf("%d: [%02d] %q\n", depth, offset, head)
		switch depth {
		case 1:
			switch string(head) {
			case "/bar":
				return MatchResult{Node: 1}
			case "/baz":
				return MatchResult{Node: 2}
			case "/foo":
				parent = 3
				continue Route
			default:
				break Route
			}
		case 2:
			switch parent {
			case 3:
				switch string(head) {
				case "/list":
					return MatchResult{Node: 4}
				default:
					parent = 5
					args = append(args, MatchArgument{Key: "name", Value: head[1:]})
					continue Route
				}
			default:
				break Route
			}
		case 3:
			switch parent {
			case 5:
				switch string(head) {
				case "/update":
					return MatchResult{Node: 6, Args: args}
				case "/delete":
					return MatchResult{Node: 7, Args: args}
				}
			}
		default:
			break Route
		}
	}
	return MatchResult{NotFound: true}
}

func TestNewGenerator(t *testing.T) {
	var g GeneratedRouter

	for k, v := range map[string]MatchResult{
		"":     {NotFound: true},
		"/bar": {Node: 1},
		"/baz": {Node: 2},
		"/foo/alex/update": {Node: 6, Args: []MatchArgument{
			{
				Key:   "name",
				Value: []byte("alex"),
			},
		}},
	} {
		t.Run(k, func(t *testing.T) {
			require.Equal(t, v, g.Match([]byte(k), nil))
		})
	}
	/*
		g.Match([]byte("/bar"))
		g.Match([]byte("/baz"))
		g.Match([]byte("/foo/{name}/delete"))
		g.Match([]byte("/foo/{name}/update"))
		g.Match([]byte("/foo/{name}"))
		g.Match([]byte("/foo/list"))
	*/
}

func BenchmarkNewGenerator(b *testing.B) {
	var g GeneratedRouter
	arg := make([]MatchArgument, 0, 5)
	p := []byte("/foo/alex/update")

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := g.Match(p, arg)
		if r.NotFound || len(r.Args) != 1 {
			b.Fatal("not found")
		}
	}
}

func TestRouter_Graph(t *testing.T) {
	var r Router
	r.Register(Route{Method: http.MethodGet, Operation: "getUser", Path: []*ir.PathPart{}})
	r.Register(Route{Method: http.MethodGet, Operation: "getUserInfo", Path: []*ir.PathPart{}})
	r.Register(Route{Method: http.MethodGet, Operation: "listUsers", Path: []*ir.PathPart{}})
	r.Register(Route{Method: http.MethodDelete, Operation: "deleteUser", Path: []*ir.PathPart{}})
	r.Register(Route{Method: http.MethodGet, Operation: "default", Path: []*ir.PathPart{}})

	require.NoError(t, r.Graph())

	for _, m := range r.Methods {
		fmt.Println(m.Method)
		for _, e := range m.Edges {
			printEdge(2, e)
		}
	}
}
