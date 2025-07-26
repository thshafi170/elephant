package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Data struct {
	NoDisplay      bool
	Hidden         bool
	Terminal       bool
	Action         string
	Name           string
	Comment        string
	Path           string
	Parent         string
	GenericName    string
	StartupWMClass string
	Icon           string
	Categories     []string
	OnlyShowIn     []string
	NotShowIn      []string
	Keywords       []string
}

func parseFile(path, l, ll string) *DesktopFile {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error(Name, "parseFile", err)
		os.Exit(1)
	}

	parts := splitIntoParsebles(data)

	f := &DesktopFile{}

	for i, v := range parts {
		data := parseData(v, l, ll)

		if i == 0 {
			f.Data = data
		} else {
			f.Actions = append(f.Actions, data)
		}
	}

	for k, v := range f.Actions {
		if len(v.Categories) == 0 {
			f.Actions[k].Categories = f.Categories
		}

		if v.Comment == "" {
			f.Actions[k].Comment = f.Comment
		}

		if v.GenericName == "" {
			f.Actions[k].GenericName = f.GenericName
		}

		f.Actions[k].Parent = f.Name
		f.Actions[k].Hidden = f.Hidden
		f.Actions[k].NoDisplay = f.NoDisplay
		f.Actions[k].NotShowIn = f.NotShowIn
		f.Actions[k].NoDisplay = f.NoDisplay
		f.Actions[k].Path = f.Path
		f.Actions[k].Terminal = f.Terminal
		f.Actions[k].StartupWMClass = f.StartupWMClass

		if len(v.Keywords) == 0 {
			f.Actions[k].Keywords = f.Keywords
		}

		if v.Icon == "" {
			f.Actions[k].Icon = f.Icon
		}
	}

	return f
}

func parseData(in []byte, l, ll string) Data {
	res := Data{}

	for line := range bytes.Lines(in) {
		line = bytes.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		switch {

		case bytes.HasPrefix(line, []byte("Keywords=")):
			res.Keywords = strings.Split(string(bytes.TrimPrefix(line, []byte("Keywords="))), ";")
		case bytes.HasPrefix(line, fmt.Appendf(nil, "Keywords[%s]=", l)):
			res.Keywords = strings.Split(string(bytes.TrimPrefix(line, fmt.Appendf(nil, "Keywords[%s]=", l))), ";")
		case bytes.HasPrefix(line, fmt.Appendf(nil, "Keywords[%s]=", ll)):
			res.Keywords = strings.Split(string(bytes.TrimPrefix(line, fmt.Appendf(nil, "Keywords[%s]=", ll))), ";")

		case bytes.HasPrefix(line, []byte("GenericName=")):
			res.GenericName = string(bytes.TrimPrefix(line, []byte("GenericName=")))
		case bytes.HasPrefix(line, fmt.Appendf(nil, "GenericName[%s]=", l)):
			res.GenericName = string(bytes.TrimPrefix(line, fmt.Appendf(nil, "GenericName[%s]=", l)))
		case bytes.HasPrefix(line, fmt.Appendf(nil, "GenericName[%s]=", ll)):
			res.GenericName = string(bytes.TrimPrefix(line, fmt.Appendf(nil, "GenericName[%s]=", ll)))

		case bytes.HasPrefix(line, []byte("Name=")):
			res.Name = string(bytes.TrimPrefix(line, []byte("Name=")))
		case bytes.HasPrefix(line, fmt.Appendf(nil, "Name[%s]=", l)):
			res.Name = string(bytes.TrimPrefix(line, fmt.Appendf(nil, "Name[%s]=", l)))
		case bytes.HasPrefix(line, fmt.Appendf(nil, "Name[%s]=", ll)):
			res.Name = string(bytes.TrimPrefix(line, fmt.Appendf(nil, "Name[%s]=", ll)))

		case bytes.HasPrefix(line, []byte("Comment=")):
			res.Comment = string(bytes.TrimPrefix(line, []byte("Comment=")))
		case bytes.HasPrefix(line, fmt.Appendf(nil, "Comment[%s]=", l)):
			res.Comment = string(bytes.TrimPrefix(line, fmt.Appendf(nil, "Comment[%s]=", l)))
		case bytes.HasPrefix(line, fmt.Appendf(nil, "Comment[%s]=", ll)):
			res.Comment = string(bytes.TrimPrefix(line, fmt.Appendf(nil, "Comment[%s]=", ll)))

		case bytes.HasPrefix(line, []byte("NoDisplay=")):
			res.NoDisplay = strings.ToLower(string(bytes.TrimPrefix(line, []byte("NoDisplay=")))) == "true"
		case bytes.HasPrefix(line, []byte("Hidden=")):
			res.Hidden = strings.ToLower(string(bytes.TrimPrefix(line, []byte("Hidden=")))) == "true"
		case bytes.HasPrefix(line, []byte("Terminal=")):
			res.Terminal = strings.ToLower(string(bytes.TrimPrefix(line, []byte("Terminal=")))) == "true"
		case bytes.HasPrefix(line, []byte("Path=")):
			res.Path = string(bytes.TrimPrefix(line, []byte("Path=")))

		case bytes.HasPrefix(line, []byte("StartupWMClass=")):
			res.StartupWMClass = string(bytes.TrimPrefix(line, []byte("StartupWMClass=")))

		case bytes.HasPrefix(line, []byte("Icon=")):
			res.Icon = string(bytes.TrimPrefix(line, []byte("Icon=")))

		case bytes.HasPrefix(line, []byte("Categories=")):
			res.Categories = strings.Split(string(bytes.TrimPrefix(line, []byte("Categories="))), ";")

		case bytes.HasPrefix(line, []byte("OnlyShowIn=")):
			res.OnlyShowIn = strings.Split(string(bytes.TrimPrefix(line, []byte("OnlyShowIn="))), ";")

		case bytes.HasPrefix(line, []byte("NotShowIn=")):
			res.NotShowIn = strings.Split(string(bytes.TrimPrefix(line, []byte("NotShowIn="))), ";")

		case bytes.Contains(line, []byte("[Desktop Action ")):
			res.Action = string(bytes.TrimPrefix(line, []byte("[Desktop Action ")))
			res.Action = strings.TrimSuffix(res.Action, "]")
		}
	}

	return res
}

func splitIntoParsebles(in []byte) [][]byte {
	actions := bytes.Contains(in, []byte("Desktop Action"))

	if !actions {
		return [][]byte{in}
	}

	parts := [][]byte{}
	posGeneric := bytes.Index(in, []byte("Desktop Entry"))
	posAction := bytes.Index(in, []byte("Desktop Action")) - 1

	parts = append(parts, in[posGeneric:posAction])

	rest := in[posAction:]

	for i := bytes.LastIndex(rest, []byte("Desktop Action")) - 1; i > 1; i = bytes.LastIndex(rest, []byte("Desktop Action")) - 1 {
		parts = append(parts, rest[i:])
		rest = rest[:i]
	}

	parts = append(parts, rest)

	return parts
}
