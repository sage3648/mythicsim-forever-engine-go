package feralbear

import (
	"testing"
	"time"
)

// Barkskin is an instant, so it leaves the melee swing alone, as an instant Maelstrom Weapon bolt does
// (docs/mythicsim-patches.md, patch 17). This is a judgment call: the Barkskin row carries
// InterruptFlags 8, the bit whose removal made Faerie Fire stop resetting the swing in 70170.
func TestBarkskinDoesNotMoveTheNextSwing(t *testing.T) {
	for _, wait := range []time.Duration{200 * time.Millisecond, 900 * time.Millisecond, 1700 * time.Millisecond} {
		sim, d := newBearInForm(t)
		// A swing has just landed, so the next one is a full weapon speed away.
		d.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
		sim.CurrentTime += wait
		d.GCD.Set(sim.CurrentTime)

		before := d.AutoAttacks.MainhandSwingAt()
		if before <= sim.CurrentTime {
			t.Fatalf("wait %v: the next swing at %v is not ahead of %v", wait, before, sim.CurrentTime)
		}
		if !d.Barkskin.Cast(sim, d.CurrentTarget) {
			t.Fatalf("wait %v: Barkskin failed to cast", wait)
		}
		if got := d.AutoAttacks.MainhandSwingAt(); got != before {
			t.Errorf("wait %v: Barkskin moved the next main hand swing from %v to %v", wait, before, got)
		}
	}
}
