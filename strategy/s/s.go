package main

import "fmt"

func main() {
	s := "9"
	m := ""
	if len(s) == 1 {
		s = "0" + s
	}
	for i, d := range s {
		if i == (len(s) - 2) {
			if i == 0 {
				m += "0"
			}
			m += "."
		}
		m += string(d)
	}
	fmt.Printf("%v \n", m)
}
