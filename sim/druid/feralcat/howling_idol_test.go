package feralcat

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
)

// Howling Idol (272427): "Reduces the cooldown of your Shifting Power ability by 1 sec" (spell 1291059).
// With the idol, Shifting Power comes back a second sooner, and a cast starts that shorter cooldown.
// Without it, the cooldown is the one the talents give.
func TestHowlingIdolCutsShiftingPowerCooldown(t *testing.T) {
	setup := func(idol bool) (*core.Simulation, *druid.Druid) {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		if idol {
			items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: 272427}
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 1},
			Raid: &proto.Raid{Parties: []*proto.Party{{Buffs: &proto.PartyBuffs{}, Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf, TalentsString: DefaultTalents,
				Equipment: &proto.EquipmentSpec{Items: items}, Buffs: &proto.IndividualBuffs{}, Spec: DefaultSpecOptions,
				Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			}}}}},
			Encounter: core.MakeSingleTargetEncounter(0),
		}, simsignals.CreateSignals())
		sim.Reset()
		return sim, sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	}

	_, bare := setup(false)
	if bare.ShiftingPower == nil {
		t.Fatal("the build has no Shifting Power")
	}
	sim, cat := setup(true)
	if got, want := cat.ShiftingPower.CD.Duration, bare.ShiftingPower.CD.Duration-time.Second; got != want {
		t.Fatalf("with the idol Shifting Power's cooldown is %v, want %v", got, want)
	}

	if !cat.InForm(druid.Cat) {
		t.Fatal("the Cat druid does not start in Cat Form")
	}
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Shifting Power did not cast")
	}
	if got, want := cat.ShiftingPower.CD.TimeToReady(sim), cat.ShiftingPower.CD.Duration; got != want {
		t.Errorf("the cast started a cooldown of %v, want the idol's %v", got, want)
	}
}
