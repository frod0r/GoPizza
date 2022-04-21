package main

import (
	"fmt"
	"testing"
)

func TestPartSub(t *testing.T) {
	n := 6
	l := 2
	values := make([]int, n)
	for i := range values {
		values[i] = i + 1
	}
	parts := partSub(values, l)
	fmt.Println("Results:")
	for _, part := range parts {
		fmt.Printf("%v\n", part)
	}
	expected := [][][]int{{{1, 2}, {3, 4}, {5, 6}},
		{{1, 2}, {3, 5}, {4, 6}},
		{{1, 2}, {3, 6}, {4, 5}},
		{{1, 3}, {2, 4}, {5, 6}},
		{{1, 3}, {2, 5}, {4, 6}},
		{{1, 3}, {2, 6}, {4, 5}},
		{{1, 4}, {2, 3}, {5, 6}},
		{{1, 4}, {2, 5}, {3, 6}},
		{{1, 4}, {2, 6}, {3, 5}},
		{{1, 5}, {2, 3}, {4, 6}},
		{{1, 5}, {2, 4}, {3, 6}},
		{{1, 5}, {2, 6}, {3, 4}},
		{{1, 6}, {2, 3}, {4, 5}},
		{{1, 6}, {2, 4}, {3, 5}},
		{{1, 6}, {2, 5}, {3, 4}}}
	for i, part := range parts {
		for j, list := range part {
			for k, elem := range list {
				if elem != expected[i][j][k] {
					t.Fail()
				}
			}
		}
	}
}
func TestPartSubNoValidation(t *testing.T) {
	n := 7
	l := 3
	values := make([]int, n)
	for i := range values {
		values[i] = i + 1
	}
	parts := partSub(values, l)
	fmt.Println("Results:")
	for _, part := range parts {
		fmt.Printf("%v\n", part)
	}

}

func TestCompromiseForNoPrefs(t *testing.T) {
	c := compromiseFor([]int{0, 1})
	fmt.Printf("%v", c)
}

func TestCompromiseFor(t *testing.T) {
	toppingsTheyLike = make(map[int]map[string]bool)
	fmt.Printf("Toppings:\n%v\n", toppings)
	userA := 1337
	aDislikes := []string{toppings[1], toppings[3]}
	toppingsTheyLike[userA] = updatePrefs(toppingsTheyLike[userA], aDislikes...)
	fmt.Printf("Toppings A likes:\n%v\n", toppingsTheyLike[userA])
	compromise := compromiseFor([]int{userA, 1})
	fmt.Printf("Compromise:\n%v\n", compromise)
	for _, c := range compromise {
		if stringIn(c, aDislikes) {
			t.Fail()
		}
	}
}

func stringIn(v string, tuple []string) bool {
	for _, t := range tuple {
		if v == t {
			return true
		}
	}
	return false
}

func TestAllCompromisesFor(t *testing.T) {
	toppingsTheyLike = make(map[int]map[string]bool)
	firstNames = make(map[int]string)
	fmt.Printf("Toppings:\n%v\n", toppings)
	userA := 1337
	firstNames[userA] = "Alice"
	aDislikes := []string{toppings[1], toppings[3]}
	toppingsTheyLike[userA] = updatePrefs(toppingsTheyLike[userA], aDislikes...)
	userB := 4242
	firstNames[userB] = "Bob"
	bDislikes := []string{toppings[1], toppings[3], toppings[5]}
	toppingsTheyLike[userB] = updatePrefs(toppingsTheyLike[userB], bDislikes...)
	userC := 1042
	cDislikes := []string{toppings[1], toppings[2], toppings[5]}
	toppingsTheyLike[userC] = updatePrefs(toppingsTheyLike[userC], cDislikes...)
	userD := 7657
	firstNames[userD] = "Dora"
	dDislikes := []string{toppings[2], toppings[5], toppings[10]}
	toppingsTheyLike[userD] = updatePrefs(toppingsTheyLike[userD], dDislikes...)
	userE := 1234
	firstNames[userE] = "Emma"
	eDislikes := []string{toppings[7], toppings[10], toppings[12]}
	toppingsTheyLike[userE] = updatePrefs(toppingsTheyLike[userE], eDislikes...)
	userF := 1197
	firstNames[userF] = "Frodo"
	fDislikes := []string{toppings[1], toppings[2], toppings[3]}
	toppingsTheyLike[userF] = updatePrefs(toppingsTheyLike[userF], fDislikes...)
	IDs := []int{userA, userB, userC, userD, userE, userF}
	numberOfPizzas := 2
	compromises := allCompromises(numberOfPizzas, IDs)
	for _, compromise := range compromises {
		sum := 0
		for _, toppings := range compromise.toppings {
			sum += len(toppings)
		}
		fmt.Printf("Compromise:\n%v share a pizzas with the %v toppings\n%v\n", compromise.participants, sum, compromise.toppings)
	}
	d := decide(numberOfPizzas, IDs)
	sum := 0
	for _, toppings := range d.toppings {
		sum += len(toppings)
	}
	fmt.Printf("Decided for %v toppings\n%v\n", sum, d)
	a := announceDecision(d)
	fmt.Println(a)
}

func TestAllCompromisesPickyEater(t *testing.T) {
	toppingsTheyLike = make(map[int]map[string]bool)
	firstNames = make(map[int]string)
	fmt.Printf("Toppings:\n%v\n", toppings)
	userA := 1337
	firstNames[userA] = "Alice"
	aDislikes := []string{toppings[1], toppings[3]}
	toppingsTheyLike[userA] = updatePrefs(toppingsTheyLike[userA], aDislikes...)
	userB := 4242
	firstNames[userB] = "Bob"
	bDislikes := toppings
	toppingsTheyLike[userB] = updatePrefs(toppingsTheyLike[userB], bDislikes...)
	userC := 1042
	cDislikes := []string{toppings[1], toppings[2], toppings[5]}
	toppingsTheyLike[userC] = updatePrefs(toppingsTheyLike[userC], cDislikes...)
	userD := 7657
	firstNames[userD] = "Dora"
	dDislikes := []string{toppings[2], toppings[5], toppings[10]}
	toppingsTheyLike[userD] = updatePrefs(toppingsTheyLike[userD], dDislikes...)
	userE := 1234
	firstNames[userE] = "Emma"
	eDislikes := []string{toppings[7], toppings[10], toppings[12]}
	toppingsTheyLike[userE] = updatePrefs(toppingsTheyLike[userE], eDislikes...)
	userF := 1197
	firstNames[userF] = "Frodo"
	fDislikes := []string{toppings[1], toppings[2], toppings[3]}
	toppingsTheyLike[userF] = updatePrefs(toppingsTheyLike[userF], fDislikes...)
	IDs := []int{userA, userB, userC, userD, userE, userF}
	numberOfPizzas := 2
	compromises := allCompromises(numberOfPizzas, IDs)
	for _, compromise := range compromises {
		sum := 0
		for _, toppings := range compromise.toppings {
			sum += len(toppings)
		}
		fmt.Printf("Compromise:\n%v share a pizzas with the %v toppings\n%v\n", compromise.participants, sum, compromise.toppings)
	}
	d := decide(numberOfPizzas, IDs)
	sum := 0
	for _, toppings := range d.toppings {
		sum += len(toppings)
	}
	fmt.Printf("Decided for %v toppings\n%v\n", sum, d)
	a := announceDecision(d)
	fmt.Println(a)
}

func TestAllCompromisesLonely(t *testing.T) {
	var ids []int
	d := decide(4, ids)
	a := announceDecision(d)
	fmt.Println(a)
}
