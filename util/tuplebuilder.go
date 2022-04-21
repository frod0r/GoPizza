package util

func tupleBuilder(values []int, depth int) (result [][]int) {
	result = make([][]int, 0)
	if depth <= 0 {
		return
	}
	nextLevelValues := make([]int, len(values))
	copy(nextLevelValues, values)
	for _, i := range values {
		//python: next_level_values.remove(i), since we iterate over elements of values, and remove the current, without manipulating
		// nextLevelValues further, it is more efficient to simply slice here.
		nextLevelValues = nextLevelValues[1:] // alternatively we could also use the index returned by the range expression.
		if len(nextLevelValues) < depth-1 {   //not cap(nextLevelValues), reslicing changes the length but not the capacity!
			continue
		}
		nextLevelTuples := tupleBuilder(nextLevelValues, depth-1)
		if nextLevelTuples != nil && len(nextLevelTuples) > 0 {
			for _, tup := range nextLevelTuples {
				result = append(result, append([]int{i}, tup...))
			}
		} else {
			result = append(result, []int{i})
		}
	}
	return
}
