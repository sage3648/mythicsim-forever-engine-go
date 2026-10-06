package mage

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// The half of a Goblin Sapper Charge that hits the thrower is a Fire spell with the spell damage proc
// mask, so its crit reaches Ignite with the mage as the target. The mage carries no Ignite dot, and the
// handler used to panic on it. Ignite now only answers crits on an enemy; the same spell's crit on the
// target still ignites it.
func TestIgniteIgnoresSapperCritOnTheMage(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{Buffs: &proto.PartyBuffs{}, Players: []*proto.Player{{
			Name: "Mage", Class: proto.Class_ClassMage, Race: proto.Race_RaceGnome, TalentsString: FireTalents,
			Equipment: &proto.EquipmentSpec{}, Buffs: &proto.IndividualBuffs{},
			Consumables: &proto.ConsumesSpec{GoblinSapper: true},
			Profession1: proto.Profession_Engineering,
			Spec:        &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{}}}},
			Rotation:    &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}}}}},
		Encounter: core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()

	mage := sim.Raid.Parties[0].Players[0].(MageAgent).GetMage()
	if mage.Talents.Ignite == 0 || mage.Ignite == nil {
		t.Fatal("FireTalents no longer take Ignite; pick a build that does")
	}
	sapperSelf := mage.GetSpell(core.GoblinSapperActionID.WithTag(1))
	if sapperSelf == nil {
		t.Fatal("no Goblin Sapper Charge self damage spell on an engineer with the charge")
	}

	crit := func(target *core.Unit) {
		result := sapperSelf.CalcDamage(sim, target, 600, sapperSelf.OutcomeAlwaysHit)
		result.Outcome = core.OutcomeCrit
		sapperSelf.DealDamage(sim, result)
	}

	crit(&mage.Unit)
	if dot := mage.Ignite.Dot(mage.CurrentTarget); dot.IsActive() {
		t.Error("a sapper crit on the mage ignited the target")
	}

	crit(mage.CurrentTarget)
	if dot := mage.Ignite.Dot(mage.CurrentTarget); !dot.IsActive() {
		t.Error("a Fire spell crit on the target no longer ignites it")
	}
}
