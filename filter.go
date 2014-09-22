package main

import (
	"regexp"
	"sync"

	"github.com/SlyMarbo/rss"
)

type Filters struct {
	filters      map[string]*regexp.Regexp
	filtersMutex *sync.RWMutex
}

func NewFilters() *Filters {
	filters := new(Filters)
	filters.filters = make(map[string]*regexp.Regexp)
	filters.filtersMutex = new(sync.RWMutex)

	return filters
}

func (filters *Filters) Match(item *rss.Item, filter string) (bool, error) {
	regex, err := filters.getFilter(filter)
	if err != nil {
		return false, err
	}

	match := regex.Match([]byte(item.Title))

	return match, nil
}

func (filters *Filters) getFilter(filter string) (*regexp.Regexp, error) {
	filters.filtersMutex.RLock()
	regex, exists := filters.filters[filter]
	filters.filtersMutex.RUnlock()

	if exists {
		return regex, nil
	}

	regex, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	filters.filtersMutex.Lock()
	filters.filters[filter] = regex
	filters.filtersMutex.Unlock()

	return regex, nil
}
