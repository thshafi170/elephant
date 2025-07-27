package common

import (
	"fmt"
	"strconv"
	"strings"
)

type Entry struct {
	Identifier string
	Text       string
	SubText    string
	Icon       string
	Provider   string
	Score      int
	Fuzzy      *FuzzyMatchInfo
}

type FuzzyMatchInfo struct {
	Field string
	Pos   *[]int
	Start int
}

func (e Entry) String() string {
	var start int
	var field string

	positions := []string{}

	if e.Fuzzy != nil {
		if e.Fuzzy.Pos != nil {
			for _, num := range *e.Fuzzy.Pos {
				positions = append(positions, strconv.Itoa(num))
			}
		}

		start = e.Fuzzy.Start
		field = e.Fuzzy.Field
	}

	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;%d;%s", e.Identifier, e.Text, e.SubText, e.Icon, e.Provider, strings.Join(positions, ","), start, field)
}
