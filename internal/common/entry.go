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
	Fuzzy      FuzzyMatchInfo
}

type FuzzyMatchInfo struct {
	Field string
	Pos   *[]int
	Start int
}

func (e Entry) String() string {
	positions := make([]string, len(*e.Fuzzy.Pos))

	for i, num := range *e.Fuzzy.Pos {
		positions[i] = strconv.Itoa(num)
	}

	return fmt.Sprintf("%s;%s;%s;%s;%s;%s;%d;%s", e.Identifier, e.Text, e.SubText, e.Icon, e.Provider, strings.Join(positions, ","), e.Fuzzy.Start, e.Fuzzy.Field)
}
