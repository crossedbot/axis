package models

import (
	"fmt"
)

const (
	TextMatchExact TextMatch = iota
	TextMatchIExact
	TextMatchPartial
	TextMatchIPartial
)

type TextMatch int

var TextMatchStrings = []string{
	"exact",
	"iexact",
	"partial",
	"ipartial",
}

func (tm TextMatch) String() (match string) {
	if int(tm) > -1 && len(TextMatchStrings) > int(tm) {
		match = TextMatchStrings[tm]
	}
	return
}

func ToTextMatch(m string) (TextMatch, error) {
	for i, tm := range TextMatchStrings {
		if tm == m {
			return TextMatch(i), nil
		}
	}
	return TextMatch(-1), fmt.Errorf("unkown text matching string")
}
