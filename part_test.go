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
