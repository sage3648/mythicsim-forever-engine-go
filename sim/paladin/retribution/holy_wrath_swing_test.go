package retribution

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

const holyWrathSpellID = 10318 // rank 2: a 2 sec hard cast in Forever

// A Retribution paladin with its melee swing running. Arcanite Reaper swings every 3.8 sec.
func newSwingingPaladin(t *testing.T) (*core.Simulation, *paladin.Paladin) {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name: "Ret", Class: proto.Class_ClassPaladin, Race: proto.Race_RaceHuman,
			TalentsString: RetTalents, Equipment: WeaponOnly, Spec: DefaultOptions,
			Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pal.AutoAttacks.EnableAutoSwing(sim)
	return sim, pal
}

// Casts Holy Wrath now and returns when it completes.
func castHolyWrath(t *testing.T, sim *core.Simulation, pal *paladin.Paladin) time.Duration {
	t.Helper()
	spell := pal.GetSpell(core.ActionID{SpellID: holyWrathSpellID})
	if !spell.Cast(sim, pal.CurrentTarget) {
		t.Fatal("Holy Wrath did not cast")
	}
	return sim.CurrentTime + spell.CurCast.CastTime
}

// Holy Wrath is a hard cast, so it holds the melee swing like Lightning Bolt does: a swing that
// comes due while it casts lands as it completes, and a cast that completes first leaves the next
// swing a full swing after it. The paladin is not a shaman, so this is the shared cast path at work.
func TestHolyWrathHoldsTheSwing(t *testing.T) {
	t.Run("swing due during the cast lands as it completes", func(t *testing.T) {
		sim, pal := newSwingingPaladin(t)
		speed := pal.AutoAttacks.MainhandSwingSpeed()
		pal.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+time.Second-speed) // due one second in
		castEnd := castHolyWrath(t, sim, pal)
		if got := pal.AutoAttacks.MainhandSwingAt(); got < castEnd || got > castEnd+time.Millisecond {
			t.Errorf("the swing due during the cast is at %v, want as the cast completes at %v", got, castEnd)
		}
	})

	t.Run("cast completes first resets the timer", func(t *testing.T) {
		sim, pal := newSwingingPaladin(t)
		speed := pal.AutoAttacks.MainhandSwingSpeed()
		pal.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime) // a swing just landed
		castEnd := castHolyWrath(t, sim, pal)
		if castEnd >= speed {
			t.Fatalf("the cast ends at %v, not before the swing at %v", castEnd, speed)
		}
		if got := pal.AutoAttacks.MainhandSwingAt(); got != castEnd+speed {
			t.Errorf("the next swing is at %v, want a full swing after the cast: %v", got, castEnd+speed)
		}
	})

	t.Run("an explicit pause stays where it is", func(t *testing.T) {
		sim, pal := newSwingingPaladin(t)
		pal.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+30*time.Second)
		before := pal.AutoAttacks.MainhandSwingAt()
		castHolyWrath(t, sim, pal)
		if got := pal.AutoAttacks.MainhandSwingAt(); got != before {
			t.Errorf("a cast moved the paused swing from %v to %v", before, got)
		}
	})

	t.Run("a cast that does not start leaves the swing alone", func(t *testing.T) {
		sim, pal := newSwingingPaladin(t)
		pal.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
		castHolyWrath(t, sim, pal)
		before := pal.AutoAttacks.MainhandSwingAt()
		// The first cast is still going and its GCD is running, so a second cast fails its checks.
		if pal.GetSpell(core.ActionID{SpellID: holyWrathSpellID}).Cast(sim, pal.CurrentTarget) {
			t.Fatal("Holy Wrath cast while another was still casting")
		}
		if got := pal.AutoAttacks.MainhandSwingAt(); got != before {
			t.Errorf("a cast that failed moved the swing from %v to %v", before, got)
		}
	})
}
