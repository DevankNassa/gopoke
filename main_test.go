package main

import (
	"testing"
	"fmt"
)

func TestCleanInput(t *testing.T){
	cases := []struct{
		input string
		expected []string
		}{
			{
				input: "  hello  world  ",
				expected: []string{"hello","world"},
			},
			{
				input:"  charmander  pikachu  raichu  ",
				expected: []string{"charmander","pikachu","raichu"},
			},
			{
				input:" charmander pikachu raichu ",
				expected: []string{"charmander","pikachu","raichu"},
			}, 
			{
				input:" a b c ",
				expected: []string{"a","b","c"},
			},
		}

	for _,c := range cases {


		actual := cleanInput(c.input)
		
		fmt.Printf("actual : %v %v\n", actual, len(actual))
		fmt.Printf("expected : %v %v\n", c.expected, len(c.expected))
		if len(actual) != len(c.expected){
			t.Errorf("length doesn't match")
			continue
		}
		for i := range actual{
			word := actual[i]
			fmt.Printf("actual word : %v\n", word)
			expectedWord := c.expected[i]
			if word != expectedWord{
				fmt.Printf("Failed Case %v",i)
				t.Errorf("not matched")
			}
		}
	}
}


