package common

import (
	"slices"
	"unicode"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

func init() {
	algo.Init("default")
}

func FuzzyScore(input, target string) (int, *[]int, int) {
	runes := []rune(input)
	chars := util.ToChars([]byte(target))
	res, pos := algo.FuzzyMatchV2(slices.ContainsFunc(runes, unicode.IsUpper), true, true, &chars, runes, true, nil)

	return res.Score, pos, res.Start
}
