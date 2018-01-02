package main

import "testing"

func TestHelperJson2Map__WrongSyntax(t *testing.T) {
	def := helperJson2Map(`hfgdfhgfd`)
	assertNil(t, def)
}

func TestHelperJson2Map__InvalidCharacters(t *testing.T) {
	def := helperJson2Map(`äöü"ßß".-!""@`)
	assertNil(t, def)
}
