package api

import "sort"

func strInList(sortedlist []string, val string)(bool) {
	ix := sort.SearchStrings(sortedlist, val)
	if ix >= len(sortedlist) {
		return false
	} else if sortedlist[ix] == val {
		return true
	}
	return false
}

