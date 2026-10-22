package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// The client rows the measured coefficient overrides replace, as they stood when each was measured. A
// client build that moves one should prompt a new measurement rather than be shadowed by the old
// constant.
func TestTheRowsMeasuredOverridesReplaceHaveNotMoved(t *testing.T) {
	for _, tc := range []struct {
		spell    int32
		position int
		coeff    float32
		site     string
	}{
		{1510, 1, 0.03, "Volley, rank 1 channel dummy (0.056 a tick)"},
		{14295, 1, 0.03, "Volley, rank 3 channel dummy (0.056 a tick)"},
	} {
		if row := float32(spelldata.MustFind(tc.spell).EffectN(tc.position).Coeff()); row != tc.coeff {
			t.Errorf("%s: the client row now carries %v; measure the override again", tc.site, row)
		}
	}
}
