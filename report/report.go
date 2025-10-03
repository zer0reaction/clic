// This file contains implementation of error reporting.

package report

import (
	"fmt"
	"os"
)

type Form struct {
	Tag    ReportTag
	File   string
	Line   uint
	Column uint
	Msg    string
}

type ReportTag uint

const (
	reportError ReportTag = iota
	ReportFatal
	ReportNonfatal
)

var errorsOccured = false

func Report(f Form) {
	errorsOccured = true

	if f.Line == 0 || f.Column == 0 {
		panic("line or column not set in report")
	}

	switch f.Tag {
	case ReportFatal:
		fmt.Printf("%s:%d:%d: fatal: %s\n",
			f.File, f.Line, f.Column, f.Msg)
		os.Exit(1)
	case ReportNonfatal:
		fmt.Printf("%s:%d:%d: error: %s\n",
			f.File, f.Line, f.Column, f.Msg)
	default:
		panic("unexpected report tag")
	}
}

func ErrorsOccured() bool {
	return errorsOccured
}
