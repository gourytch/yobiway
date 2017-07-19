package yobit

import "strings"

func parse_yobit_pairname(s string) (pairname, token, currency string) {
	V := strings.Split(strings.ToUpper(s), "_")
	return V[0] + "/" + V[1], V[0], V[1]
}
