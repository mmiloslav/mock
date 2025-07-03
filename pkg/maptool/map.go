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

func UnsortJSONMap(s []SortedJSONMap) map[string][]string {
	if len(s) == 0 {
		return nil
	}

	m := make(map[string][]string, len(s))
	for _, item := range s {
		m[item.Key] = item.Values
	}

	return m
}
