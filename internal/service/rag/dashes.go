package rag

import "regexp"

var (
	numericRangeDash = regexp.MustCompile(`(\d)\s*[\x{2013}\x{2014}]\s*(\d)`)
	lineLeadingDash  = regexp.MustCompile(`(?m)^([ \t]*)[\x{2013}\x{2014}][ \t]*`)
	inlineDash       = regexp.MustCompile(`[ \t]*[\x{2013}\x{2014}][ \t]*`)
	doubledComma     = regexp.MustCompile(`,(\s*,)+`)
	commaBeforeStop  = regexp.MustCompile(`,\s*([.!?;:)])`)
	commaAtLineEnd   = regexp.MustCompile(`(?m),[ \t]*$`)
)

// normalizeDashes rewrites em and en dashes the model may still emit despite
// the system prompt, so answers read as plain conversational text.
func normalizeDashes(text string) string {
	text = numericRangeDash.ReplaceAllString(text, "$1-$2")
	text = lineLeadingDash.ReplaceAllString(text, "$1- ")
	text = inlineDash.ReplaceAllString(text, ", ")
	text = doubledComma.ReplaceAllString(text, ",")
	text = commaBeforeStop.ReplaceAllString(text, "$1")
	text = commaAtLineEnd.ReplaceAllString(text, "")
	return text
}
