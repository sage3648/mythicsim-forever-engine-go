package paladin

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
		{679, 1, 0.429, "Holy Strike, rank 1 (patch 143: 0.286)"},
		{10333, 1, 0.429, "Holy Strike, rank 8 (patch 143: 0.286)"},
		{20925, 2, 0.08, "Holy Shield, rank 1 (patch 143: 0.07)"},
		{20928, 2, 0.08, "Holy Shield, rank 3 (patch 143: 0.07)"},
		{20187, 1, 0.429, "Judgement of Righteousness, rank 1 (patch 143: 0.70)"},
		{20286, 1, 0.429, "Judgement of Righteousness, rank 8 (patch 143: 0.70)"},
		{20231, 1, 0.1, "Seal of Fury, per hit (patch 143: 0.15)"},
		{1311647, 1, 0.1, "Seal of Fury, per hit (patch 143: 0.15)"},
		{20183, 1, 0.429, "Judgement of Fury (patch 143: 0.67)"},
		{1311655, 1, 0.429, "Judgement of Fury (patch 143: 0.67)"},
	} {
		if row := float32(spelldata.MustFind(tc.spell).EffectN(tc.position).Coeff()); row != tc.coeff {
			t.Errorf("%s: the client row now carries %v; measure the override again", tc.site, row)
		}
	}
}
