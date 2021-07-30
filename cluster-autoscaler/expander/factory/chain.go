package factory

import (
	"k8s.io/autoscaler/cluster-autoscaler/expander"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type chainStrategy struct {
	filters  []expander.Filter
	fallback expander.Strategy
}

func newChainStrategy(filters []expander.Filter, fallback expander.Strategy) expander.Strategy {
	return &chainStrategy{
		filters:  filters,
		fallback: fallback,
	}
}

func (c *chainStrategy) BestOption(options []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo) *expander.Option {
	filteredOptions := options
	for _, filter := range c.filters {
		filteredOptions = filter.BestOptions(filteredOptions, nodeInfo)
		if len(filteredOptions) == 1 {
			return &filteredOptions[0]
		}
	}
	return c.fallback.BestOption(filteredOptions, nodeInfo)
}
