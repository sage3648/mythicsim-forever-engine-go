package druid

import (
	"math"
	"testing"
)

// Hameru's beta logs (7 October 2026): 99 more attack power (225 to 324) added 5 to the hit and to a
// tick, 0.0526 of attack power on both. The rank the engine casts keeps its own base.
func TestRakeAddsAttackPowerShareToHitAndTick(t *testing.T) {
	const low, high = 225.0, 324.0
	want := 0.0526 * (high - low)
	if got := rakeHitDamage(high) - rakeHitDamage(low); math.Abs(got-want) > 1e-9 {
		t.Errorf("Rake hit gains %.4f from %v more attack power, want %.4f", got, high-low, want)
	}
	if got := rakeTickDamage(high) - rakeTickDamage(low); math.Abs(got-want) > 1e-9 {
		t.Errorf("Rake tick gains %.4f from %v more attack power, want %.4f", got, high-low, want)
	}
	if rakeHitDamage(0) <= rakeTickDamage(0) || rakeTickDamage(0) <= 0 {
		t.Errorf("Rake base %v hit, %v tick: want the rank's client values", rakeHitDamage(0), rakeTickDamage(0))
	}
}
