// Package util provides general utility.
package util

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/abenz1267/elephant/internal/providers"
)

func GenerateDoc() {
	fmt.Println("# Elephant")

	fmt.Println("A service providing various datasources which can be triggered to perform actions.")
	fmt.Println()
	fmt.Println("Run `elephant -h` to get an overview of the available commandline flags and actions.")

	fmt.Println("## Provider Configuration")

	p := []providers.Provider{}

	for _, v := range providers.Providers {
		p = append(p, v)
	}

	slices.SortFunc(p, func(a, b providers.Provider) int {
		return strings.Compare(*a.NamePretty, *b.NamePretty)
	})

	for _, v := range p {
		v.PrintDoc()
	}
}

func PrintConfig(c any) {
	fmt.Println("| Field | Type | Default | Description |")
	fmt.Println("| --- | ---- | ---- | --- |")
	printStructDesc(c)
	fmt.Println()
}

func printStructDesc(c any) {
	val := reflect.ValueOf(c)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		fmt.Println("Not a struct")
		return
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if field.PkgPath == "" {
			if field.Anonymous {
				printStructDesc(fieldValue.Interface())
				continue
			}

			name := field.Tag.Get("koanf")
			fmt.Printf("|%s|%s|%s|%s|\n",
				name, field.Type, field.Tag.Get("default"), field.Tag.Get("desc"))

		}
	}
}
