package scan

import "errors"

type outputFormat string

const (
	outputFile   outputFormat = "file"
	outputQueue  outputFormat = "queue"
	outputStdOut outputFormat = "stdout"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *outputFormat) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *outputFormat) Set(v string) error {
	switch v {
	case "file", "queue", "stdout":
		*e = outputFormat(v)
		return nil
	default:
		return errors.New(`must be one of "file", "queue", or "stdout"`)
	}
}

// Type is only used in help text
func (e *outputFormat) Type() string {
	return "outputFormat"
}
