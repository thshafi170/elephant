package main

import (
	"fmt"

	"github.com/abenz1267/elephant/internal/util"
)

func PrintDoc() {
	fmt.Printf("### %s\n", NamePretty)
	fmt.Println("Provides access to all your installed desktop applications.")
	fmt.Println()
	util.PrintConfig(Config{}, Name)
}
