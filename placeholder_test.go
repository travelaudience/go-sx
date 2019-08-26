package sx_test

import (
	"testing"

	sx "github.com/travelaudience/go-sx"
)

func TestPlaceholders(t *testing.T) {

	sx.SetNumberedPlaceholders(false)

	t.Run("? placeholders", func(t *testing.T) {
		want := []string{"?", "?", "?"}
		var p sx.Placeholder

		for i, x := range want {
			y := p.Next()
			if x != y {
				t.Errorf("case a-%d: expected %s, got %s", i, x, y)
			}
		}
	})

	sx.SetNumberedPlaceholders(true)

	t.Run("numbered placeholders", func(t *testing.T) {
		want := []string{"$1", "$2", "$3"}
		var p sx.Placeholder

		for i, x := range want {
			y := p.Next()
			if x != y {
				t.Errorf("case b-%d: expected %s, got %s", i, x, y)
			}
		}
	})
}
