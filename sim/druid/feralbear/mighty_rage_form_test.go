package feralbear

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
)

// Mighty Rage Potion (13442) has no shapeshift restriction in the client, so a bear drinks it in Bear
// Form and keeps the form, and its automatic use is not held back until the bear leaves the form.
func TestBearDrinksMightyRageAndStaysInForm(t *testing.T) {
	player := &proto.Player{
		Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		Equipment: &proto.EquipmentSpec{}, Spec: DefaultSpecOptions,
		Consumables: &proto.ConsumesSpec{PotId: 13442, Potions: []int32{13442}},
		Rotation:    &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	d := sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	sim.CurrentTime = time.Second
	d.ClearForm(sim)
	if !d.BearForm.Cast(sim, d.CurrentTarget) {
		t.Fatal("Bear Form failed")
	}

	potion := d.GetMajorCooldown(core.ActionID{ItemID: 13442})
	if potion == nil {
		t.Fatal("Mighty Rage Potion is not one of the bear's cooldowns")
	}
	if potion.ShouldActivate != nil && !potion.ShouldActivate(sim, &d.Character) {
		t.Error("the automatic use of the potion is held back in Bear Form")
	}
	if !potion.Spell.Cast(sim, d.CurrentTarget) {
		t.Fatal("the potion did not cast in Bear Form")
	}
	if !d.InForm(druid.Bear) {
		t.Error("drinking the potion dropped Bear Form")
	}
	if potion.Spell.CD.IsReady(sim) {
		t.Error("drinking the potion did not start its cooldown")
	}
}
