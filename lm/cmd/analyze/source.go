package analyze

import (
	"fmt"
	"strings"
)

type Source int

const (
	SourceNone Source = iota
	SourceCLI
	SourceFile
	SourceSQS
)

var sourceName = map[Source]string{
	SourceNone: "none",
	SourceCLI:  "cli",
	SourceFile: "file",
	SourceSQS:  "sqs",
}

func (s *Source) String() string {
	return sourceName[*s]
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (s *Source) Set(v string) error {
	source, success := sourceFromName(v)
	if !success {
		return fmt.Errorf("unable to find Source enum value for %q", v)
	}
	*s = source
	return nil
}

// Type is only used in help text
func (s *Source) Type() string {
	return "source"
}

func sourceNames() []string {
	names := []string{}
	for k, v := range sourceName {
		if k == SourceNone {
			continue
		}
		names = append(names, v)
	}
	return names
}

func sourceFromName(name string) (Source, bool) {
	lname := strings.ToLower((name))
	for k, v := range sourceName {
		if k == SourceNone {
			continue
		}

		if v == lname {
			return k, true
		}
	}

	return SourceNone, false
}
