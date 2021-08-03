package factory

import (
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type substringTestFilterStrategy struct {
	substring string
}

func newSubstringTestFilterStrategy(substring string) *substringTestFilterStrategy {
	return &substringTestFilterStrategy{
		substring: substring,
	}
}

func (s *substringTestFilterStrategy) BestOptions(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) []expander.Option {
	var ret []expander.Option
	for _, option := range expansionOptions {
		if strings.Contains(option.Debug, s.substring) {
			ret = append(ret, option)
		}
	}
	return ret

}

func (s *substringTestFilterStrategy) BestOption(expansionOptions []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
	ret := s.BestOptions(expansionOptions, nodeInfo)
	if len(ret) == 0 {
		return nil
	}
	return &ret[0]
}

func (s *substringTestFilterStrategy) AlwaysUniqueOption() bool {
	return false
}

func TestChainStrategy_BestOption(t *testing.T) {
	for name, tc := range map[string]struct {
		filters  []expander.Filter
		fallback expander.Strategy
		options  []expander.Option
		expected *expander.Option
	}{
		"selects with no filters": {
			filters:  []expander.Filter{},
			fallback: newSubstringTestFilterStrategy("a"),
			options: []expander.Option{
				*newOption("b"),
				*newOption("a"),
			},
			expected: newOption("a"),
		},
		"filters with one filter": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("a"),
			},
			fallback: newSubstringTestFilterStrategy("b"),
			options: []expander.Option{
				*newOption("ab"),
				*newOption("b"),
			},
			expected: newOption("ab"),
		},
		"filters with multiple filters": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("a"),
				newSubstringTestFilterStrategy("b"),
			},
			fallback: newSubstringTestFilterStrategy("x"),
			options: []expander.Option{
				*newOption("xab"),
				*newOption("xa"),
				*newOption("x"),
			},
			expected: newOption("xab"),
		},
		"selects from multiple after filters": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("x"),
			},
			fallback: newSubstringTestFilterStrategy("a"),
			options: []expander.Option{
				*newOption("xc"),
				*newOption("xaa"),
				*newOption("xab"),
			},
			expected: newOption("xaa"),
		},
		"short circuits": {
			filters: []expander.Filter{
				newSubstringTestFilterStrategy("a"),
				newSubstringTestFilterStrategy("b"),
			},
			fallback: newSubstringTestFilterStrategy("x"),
			options: []expander.Option{
				*newOption("a"),
			},
			expected: newOption("a"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			subject := newChainStrategy(tc.filters, tc.fallback)
			actual := subject.BestOption(tc.options, nil)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func newOption(debug string) *expander.Option {
	return &expander.Option{
		Debug: debug,
	}
}
