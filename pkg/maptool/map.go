package maptool

import "sort"

type SortedJSONMap struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

// Сортирует map[string][]string по ключам и значениям (внутри слайсов)
func SortJSONMap(m map[string][]string) []SortedJSONMap {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make([]SortedJSONMap, 0, len(m))
	for _, k := range keys {
		vals := make([]string, len(m[k]))
		copy(vals, m[k])
		sort.Strings(vals)

		sorted = append(sorted, SortedJSONMap{
			Key:    k,
			Values: vals,
		})
	}

	return sorted
}
