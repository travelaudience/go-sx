package sx

import (
	"testing"
)

func TestPlaceholders(t *testing.T) {

	SetNumberedPlaceholders(false)

	t.Run("? placeholders", func(t *testing.T) {
		want := []string{"?", "?", "?"}
		var p Placeholder

		for i, x := range want {
			y := p.Next()
			if x != y {
				t.Errorf("case a-%d: expected %s, got %s", i, x, y)
			}
		}
	})

	SetNumberedPlaceholders(true)

	t.Run("numbered placeholders", func(t *testing.T) {
		want := []string{"$1", "$2", "$3"}
		var p Placeholder

		for i, x := range want {
			y := p.Next()
			if x != y {
				t.Errorf("case b-%d: expected %s, got %s", i, x, y)
			}
		}
	})
}
