package location

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	file1 = File{
		Name:   "spec.json",
		Source: "spec.json",
	}
	file2 = File{
		Name:   "user.json",
		Source: "user.json",
	}
)

func Test_chunkReports(t *testing.T) {
	highlight := func(l int) Highlight {
		return Highlight{
			Pos: Position{
				Line: l,
			},
			Color: nil,
		}
	}

	tests := []struct {
		reports []Report
		context int
		want    []reportChunk
	}{
		{
			[]Report{
				{file1, Position{Line: 5}, "Error message"},
				{file1, Position{Line: 100}, ""},
				{file1, Position{Line: 13}, ""},
				{file1, Position{Line: 7}, ""},
			},
			3,
			[]reportChunk{
				{"Error message", file1, []Highlight{
					highlight(5),
					highlight(7),
					highlight(13),
				}},
				{"", file1, []Highlight{
					highlight(100),
				}},
			},
		},
		{
			[]Report{
				{file1, Position{Line: 5}, "Error message"},
				{file1, Position{Line: 100}, ""},
				{file2, Position{Line: 6}, ""},
				{file1, Position{Line: 7}, ""},
			},
			3,
			[]reportChunk{
				{"Error message", file1, []Highlight{
					highlight(5),
					highlight(7),
				}},
				{"", file1, []Highlight{
					highlight(100),
				}},
				{"", file2, []Highlight{
					highlight(6),
				}},
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			r := chunkReports(tt.reports, tt.context, nil)
			require.Equal(t, tt.want, r)
		})
	}
}

func TestMultiError_Error(t *testing.T) {
	tests := []struct {
		reports []Report
		expect  string
	}{
		{
			reports: []Report{},
			expect:  "empty error",
		},
		{
			reports: []Report{
				{file1, Position{Line: 10}, "bad line"},
			},
			expect: "at spec.json:10: bad line",
		},
		{
			reports: []Report{
				{file1, Position{Line: 10}, "bad line"},
				{file2, Position{Line: 100}, "second bad line"},
			},
			expect: "- at spec.json:10: bad line\n" +
				"- at user.json:100: second bad line\n",
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
			me := &MultiError{
				reports: tt.reports,
			}
			require.Equal(t, me.Error(), tt.expect)
			require.EqualError(t, me, tt.expect)
		})
	}
}
