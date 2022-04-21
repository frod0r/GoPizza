package util

import (
	"pizza_decision_bot/data"
	"sort"
)

// when Compromise is unwrapped, a group of participants ([]int) corresponds to a set of toppings ([]string) of a pizza they share
type Compromise struct {
	Toppings     [][]string
	Participants [][]int
}

func Decide(numberOfPizzas int, IDs []int) Compromise {
	compromises := AllCompromises(numberOfPizzas, IDs)
	if len(compromises) == 0 {
		return Compromise{}
	}
	sort.Slice(compromises, func(i, j int) bool {
		//return people[i].Age < people[j].Age })
		sumI := 0
		for _, toppingsI := range compromises[i].Toppings {
			sumI += len(toppingsI)
		}
		sumJ := 0
		for _, toppingsJ := range compromises[j].Toppings {
			sumJ += len(toppingsJ)
		}
		return sumI > sumJ
	})
	return compromises[0]
	//for i, compromise := range compromises
}

// AllCompromises calculates all possible compromises that can be made with the given people.
func AllCompromises(numberOfPizzas int, IDs []int) (compromises []Compromise) {
	//todo add sanity checks
	sort.Ints(IDs)
	l := (len(IDs) + numberOfPizzas - 1) / numberOfPizzas // ceil of n/noOfPizzas, alternatively ```l := 1 + (len(IDs) - 1) / numberOfPizzas``` to avoid overflows, but in that case we have all other kinds of problems.
	partitions := PartSub(IDs, l)
	compromises = make([]Compromise, len(partitions)) //compromises = make([][][]string, len(partitions))
	for i, part := range partitions {
		compromises[i].Toppings = make([][]string, len(part))
		compromises[i].Participants = part
		for j, pizzaPeople := range part {
			compromises[i].Toppings[j] = CompromiseFor(pizzaPeople)
		}
	}
	return
}

/*func copySliceToMap(IDs []int) map[int]nothing {
	//~key = index value = id~ key = value, index omitted
	var mappedIDs = make(map[int]nothing, len(IDs))
	for _, value := range IDs {
		mappedIDs[value] = nothing{}
	}
	return mappedIDs
}

type nothing struct{}*/

func CompromiseFor(IDs []int) (resultToppings []string) {
	consensusOnDislikedToppings := make(map[string]bool)
	for _, id := range IDs {
		for topping, pref := range data.ToppingsTheyLike[id] {
			if !pref {
				consensusOnDislikedToppings[topping] = true
			}
		}
	}
	for _, topping := range data.Toppings {
		if !consensusOnDislikedToppings[topping] {
			resultToppings = append(resultToppings, topping)
		}
	}
	return resultToppings
}

//jakobs algo
func PartSub(values []int, partSize int) (partitions [][][]int) {
	if values == nil || len(values) == 0 {
		return
	}
	partitions = make([][][]int, 0)
	partSize = min(partSize, len(values))

	tuples := fixedFirstTupleBuilder(values[0], values, partSize)
	for _, tup := range tuples {
		remaining := calculateRemaining(values, tup)
		remPartitions := PartSub(remaining, partSize)
		if remPartitions != nil && len(remPartitions) > 0 {
			for _, part := range remPartitions {
				partitions = append(partitions, append([][]int{tup}, part...)) //todo inefficient, see https://stackoverflow.com/questions/53737435/how-to-prepend-int-to-slice and https://github.com/golang/go/wiki/SliceTricks
			}
		} else {
			partitions = append(partitions, [][]int{tup})
		}
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func calculateRemaining(values, tuple []int) []int {
	//python: remaining = [v for v in values if v not in tup]
	//result := make([]int, len(values)) //uses more memory than needed, but is maybe more efficient than append
	var result []int
	for _, v := range values {
		if !In(v, tuple) {
			result = append(result, v)
		}
	}
	return result
}

func In(v int, tuple []int) bool {
	//todo make more efficient by using int->noting maps, see copySliceToMap
	for _, t := range tuple {
		if v == t {
			return true
		}
	}
	return false
}

func fixedFirstTupleBuilder(fixed int, values []int, depth int) (tuples [][]int) {
	//python: values.remove(fixed) but only usage is called with fixed == values[0] so slicing is easier
	values = values[1:]
	depth--
	if depth == 0 {
		return [][]int{{fixed}}
	}
	suffixTuples := tupleBuilder(values, depth)
	tuples = make([][]int, 0)
	for _, tup := range suffixTuples {
		tuples = append(tuples, append([]int{fixed}, tup...))
	}
	return
}

// Own try at implementing a partition algorithm
/*func partition(IDs []int, k int) (result [][][]int) {
	//nehme an IDs ist sortiert
	n := len(IDs)
	b := combin.Binomial(n, k) / k
	l := len(IDs) / k
	result = make([][][]int, b)
	if l == 1 {
		for i := 0; i < k; i++ {
			result[i][0][0] = IDs[i]
		}
		return result
	}

	*
		for sb, id := range IDs {
				result[0][0] = id
				restResult := partition(IDs[sb:], k-1)
	*

	for sb := 0; sb < b; sb++ {
		result[sb] = make([][]int, k)
		result[sb][0][0] = IDs[0]
		result[sb][0][1] = IDs[1]
		for id, j := range IDs[0:] {
			if j < l {
				result[sb][0] = []int{id} // added afterwards without much thought, see below
			}
		}
		restResult := partition(IDs[sb:], k)
		result = append(result, restResult...) // added afterwards without much thought to make it compilable without
		// removing or commenting out this code to not get confused with the other commented out parts of code here,
		// but to be able to still see my original thoughts on this. To be removed in future commits
	}

	return
}*/

/*func oldDecide(numberOfPizzas int, IDs []int) [][][]string {
	//todo fails if number of pizzas is larger than number of people
	combinations := combin.Combinations(len(IDs), numberOfPizzas)
	//replace indices of ids with ids. todo integrate in Combinations func
	//also for every combination save the result
	results := make([][][]string, len(combinations))
	//var allOthers = make(map[uint64]nothing) todo compare if this contains already picked set of users
	for combNo, combination := range combinations {
		picked := make(map[int]nothing, len(IDs))
		others := copySliceToMap(IDs)
		for _, idIndex := range combination {
			delete(others, idIndex) // Todo: Das funktioniert so bis jetzt nur für 2 pizzen...
			//combination[i] = IDs[idIndex]
			picked[IDs[idIndex]] = nothing{}
		}

		results[combNo] = make([][]string, 2)
		results[combNo][0] = compromiseFor(picked)
		results[combNo][1] = compromiseFor(others)

	}
	return results
	//todo moment mal, ich will ja mögliche kombinationen mit allen elementen drin also quasi alle permutationen mit numberOfPizzas -1 aufteilungen dazwischen und das dann ohne duplikate...
	// siehe https://chat.stackexchange.com/transcript/message/3837894#3837894 and https://mathematica.stackexchange.com/questions/3044/partition-a-set-into-subsets-of-size-k/3050#3050
}*/
