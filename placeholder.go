package sx

import "strconv"

var numberedPlaceholders bool

// SetNumberedPlaceholders sets the style of placeholders to be used for generated queries.  If yes is true, then
// postgres-style "$n" placeholders will be used for all future queries.  If yes is false, then mysql-style "?"
// placeholders will be used.  This setting may be changed at any time.  Default is false.
func SetNumberedPlaceholders(yes bool) {
	numberedPlaceholders = yes
}

// A Placeholder is a generator for the currently selected placeholder type.  See SetNumberedPlaceholders().
type Placeholder int

// String displays the current placeholder value in its chosen format (either "?" or "$n").
func (p Placeholder) String() string {
	if numberedPlaceholders {
		return "$" + strconv.Itoa(int(p))
	}
	return "?"
}

// Next increments the placeholder value and returns the string value of the next placeholder in sequence.
//
// When using numbered placeholders, a zero-valued placeholder will return "$1" on its first call to Next().
// When using ?-style placeholders, Next always returns "?".
func (p *Placeholder) Next() string {
	*p++
	return p.String()
}
