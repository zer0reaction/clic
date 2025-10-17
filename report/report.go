// This file contains implementation of error reporting.

package report

import (
	"fmt"
	"os"
)

type Reporter struct {
	FileName   string
	errorCount uint
}

type Form struct {
	Tag    ReportTag
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

func (r *Reporter) Report(f Form) {
	// TODO: Dont increment this on warnings (when they are added)
	r.errorCount += 1

	if f.Line == 0 || f.Column == 0 {
		panic("line or column not set in report")
	}

	switch f.Tag {
	case ReportFatal:
		fmt.Printf("%s:%d:%d: fatal: %s\n",
			r.FileName, f.Line, f.Column, f.Msg)
		os.Exit(1)

	case ReportNonfatal:
		fmt.Printf("%s:%d:%d: error: %s\n",
			r.FileName, f.Line, f.Column, f.Msg)

	default:
		panic("not implemented")
	}
}

func (r *Reporter) ExitOnErrors(code int) {
	if r.errorCount > 0 {
		os.Exit(code)
	}
}
