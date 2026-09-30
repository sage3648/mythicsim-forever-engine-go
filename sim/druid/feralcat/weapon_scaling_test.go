package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
)

func TestFeralWeaponFollowsEquippedTableDPSAcrossForms(t *testing.T) {
	player := &proto.Player{Name: "Weapon scaling", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: 12783}}},
		Spec:      DefaultSpecOptions, Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL}}
	sim := core.NewSim(&proto.RaidSimRequest{SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:      core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: core.MakeSingleTargetEncounter(0)}, simsignals.CreateSignals())
	sim.Reset()
	d := sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	if w := d.GetCatWeapon(); w.BaseDamageMin != 33 || w.BaseDamageMax != 50 || w.SwingSpeed != 1 {
		t.Fatalf("Heartseeker cat weapon: %+v", w)
	}
	if w := d.GetBearWeapon(); w.BaseDamageMin != 82 || w.BaseDamageMax != 124 || w.SwingSpeed != 2.5 {
		t.Fatalf("Heartseeker bear weapon: %+v", w)
	}
	// Re-entering the form must not restore the weapon captured during setup.
	d.ClearForm(sim)
	d.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.NewItem(core.ItemSpec{ID: 11921})
	d.CatFormAura.Activate(sim)
	if w := d.AutoAttacks.MH(); w.BaseDamageMin != 39 || w.BaseDamageMax != 59 {
		t.Fatalf("cat re-entry kept the previous weapon: %+v", w)
	}
	d.ClearForm(sim)
	d.BearFormAura.Activate(sim)
	if w := d.AutoAttacks.MH(); w.BaseDamageMin != 98 || w.BaseDamageMax != 147 {
		t.Fatalf("bear re-entry weapon: %+v", w)
	}
	d.ClearForm(sim)
	d.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{}
	d.CatFormAura.Activate(sim)
	if w := d.AutoAttacks.MH(); w.BaseDamageMin != 1 || w.BaseDamageMax != 1 {
		t.Fatalf("unarmed form weapon: %+v", w)
	}
}
