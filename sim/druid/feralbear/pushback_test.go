package feralbear

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
)

const linkensBoomerang = 11905

// A tanking Bear wearing Linken's Boomerang, with the target's own swings stopped so the test places
// the one hit it takes.
func newBoomerangTank(t *testing.T) (*core.Simulation, *druid.Druid, *core.Spell) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket1+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: linkensBoomerang}

	raid := core.SinglePlayerRaidProto(&proto.Player{Name: "Bear", Class: proto.Class_ClassDruid,
		Race: proto.Race_RaceNightElf, TalentsString: DefaultTalents, Spec: DefaultSpecOptions,
		Equipment: &proto.EquipmentSpec{Items: items}, InFrontOfTarget: true, DistanceFromTarget: 10,
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL}},
		&proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	sim := core.NewSim(&proto.RaidSimRequest{SimOptions: &proto.SimOptions{RandomSeed: 1}, Raid: raid,
		Encounter: core.MakeSingleTargetEncounter(0)}, simsignals.CreateSignals())
	sim.Reset()

	d := sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	boomerang := d.GetSpell(core.ActionID{ItemID: linkensBoomerang})
	if boomerang == nil {
		t.Fatal("no Linken's Boomerang on-use")
	}
	d.PseudoStats.PushbackChance = 1
	d.CurrentTarget.AutoAttacks.CancelAutoSwing(sim)
	return sim, d, boomerang
}

func stepTo(t *testing.T, sim *core.Simulation, until time.Duration) {
	t.Helper()
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: until, OnAction: func(*core.Simulation) {}}))
	for steps := 0; sim.CurrentTime < until; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// The target's swing lands at hitAt into the boomerang's 0.5 s cast. Returns how many times the cast
// completed by 1.5 s and when it last ended.
func boomerangCastHitAt(t *testing.T, hitAt time.Duration) (int32, time.Duration) {
	t.Helper()
	sim, d, boomerang := newBoomerangTank(t)
	target := d.CurrentTarget
	start := sim.CurrentTime
	if !boomerang.Cast(sim, target) || !d.Hardcast.Pushback || d.Hardcast.Expires != start+500*time.Millisecond {
		t.Fatalf("Linken's Boomerang did not start a 0.5 s cast with pushback: casting until %v, pushback %v", d.Hardcast.Expires, d.Hardcast.Pushback)
	}

	completedAt := time.Duration(-1)
	onComplete := d.Hardcast.OnComplete
	d.Hardcast.OnComplete = func(sim *core.Simulation, target *core.Unit) {
		completedAt = sim.CurrentTime
		onComplete(sim, target)
	}

	stepTo(t, sim, start+hitAt)
	swing := target.AutoAttacks.MHAuto()
	if result := swing.CalcAndDealDamage(sim, &d.Unit, 300, swing.OutcomeAlwaysHit); !result.Landed() || result.Damage <= 0 {
		t.Fatalf("the swing at %v did not land for damage", hitAt)
	}
	stepTo(t, sim, start+1500*time.Millisecond)
	return boomerang.SpellMetrics[target.UnitIndex].Casts, completedAt - start
}

// The pushback roll runs one spell batch window (10 ms) after the hit, and used to push back whatever
// cast had its end time then. A swing 10 ms before the boomerang's 0.5 s cast completes rolled at
// 0.5 s, before the cast's completion that same instant, and moved it to 1 s. A swing a little later
// rolled after the completion, pushed back the finished cast and scheduled it again, so the boomerang
// completed twice at 0.5 s and hit twice, as in MythicSim's feral-bear-druid-boomerang-pushback-after-cast
// log. A cast that has ended by the time the roll runs is left alone.
func TestPushbackLeavesACompletedCastAlone(t *testing.T) {
	for _, hitAt := range []time.Duration{490 * time.Millisecond, 495 * time.Millisecond, 499 * time.Millisecond} {
		if casts, at := boomerangCastHitAt(t, hitAt); casts != 1 || at != 500*time.Millisecond {
			t.Errorf("a swing at %v: %d casts, the last completed at %v; want 1 at 500ms", hitAt, casts, at)
		}
	}
}

// A swing well inside the cast still pushes it back: at 0.2 s the roll runs at 0.21 s, and the cast
// moves by the 0.21 s already cast.
func TestPushbackStillDelaysACastInProgress(t *testing.T) {
	if casts, at := boomerangCastHitAt(t, 200*time.Millisecond); casts != 1 || at != 710*time.Millisecond {
		t.Errorf("a swing at 0.2 s: %d casts, the last completed at %v; want 1 at 710ms", casts, at)
	}
}
