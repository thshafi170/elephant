package common

type Entry struct {
	Identifier    string
	Text          string
	SubText       string
	Icon          string
	Provider      string
	Score         int
	FuzzyPos      *[]int
	FuzzyPosStart int
}
