package warlock

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func newStoneWarlock(t *testing.T, stone proto.WarlockOptions_WeaponStone, oil int32) *Warlock {
	t.Helper()
	player := &proto.Player{
		Name: "lock", Class: proto.Class_ClassWarlock, Race: proto.Race_RaceUndead,
		Equipment: &proto.EquipmentSpec{}, Consumables: &proto.ConsumesSpec{MhImbueId: oil},
		Spec: &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{ClassOptions: &proto.WarlockOptions{
			Summon: proto.WarlockOptions_NoSummon, Armor: proto.WarlockOptions_DemonArmor, WeaponStone: stone}}}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 7},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].(WarlockAgent).GetWarlock()
}

// Firestone (23483) adds 21 Fire damage and 2% spell crit, Spellstone (1237165) 21 Fire and Shadow
// damage and 2% cast speed, each beside Brilliant Wizard Oil (mythicsim patch 101).
func TestWeaponStonesStackWithWizardOil(t *testing.T) {
	const brilliantWizardOil = 25122
	none := newStoneWarlock(t, proto.WarlockOptions_NoWeaponStone, brilliantWizardOil)
	fire := newStoneWarlock(t, proto.WarlockOptions_Firestone, brilliantWizardOil)
	shadow := newStoneWarlock(t, proto.WarlockOptions_Spellstone, brilliantWizardOil)

	delta := func(w *Warlock, stat stats.Stat) float64 { return w.GetStat(stat) - none.GetStat(stat) }
	if got := delta(fire, stats.FireDamage); got != 21 {
		t.Errorf("Firestone fire damage %v, want 21", got)
	}
	if got := delta(fire, stats.SpellCritPercent); got != 2 {
		t.Errorf("Firestone spell crit %v, want 2", got)
	}
	if got := delta(fire, stats.ShadowDamage); got != 0 {
		t.Errorf("Firestone shadow damage %v, want 0", got)
	}
	if got := delta(shadow, stats.ShadowDamage); got != 21 {
		t.Errorf("Spellstone shadow damage %v, want 21", got)
	}
	if got := delta(shadow, stats.FireDamage); got != 21 {
		t.Errorf("Spellstone fire damage %v, want 21 (school mask 36)", got)
	}
	if got := shadow.PseudoStats.CastSpeedMultiplier / none.PseudoStats.CastSpeedMultiplier; got < 1.0199 || got > 1.0201 {
		t.Errorf("Spellstone cast speed x%v, want x1.02", got)
	}
	if got := fire.PseudoStats.CastSpeedMultiplier / none.PseudoStats.CastSpeedMultiplier; got != 1 {
		t.Errorf("Firestone cast speed x%v, want x1", got)
	}
	// The oil still counts beside either stone.
	for _, w := range []*Warlock{none, fire, shadow} {
		if w.GetStat(stats.SpellDamage) < 36 {
			t.Errorf("Brilliant Wizard Oil's 36 spell damage missing: %v", w.GetStat(stats.SpellDamage))
		}
	}
}
