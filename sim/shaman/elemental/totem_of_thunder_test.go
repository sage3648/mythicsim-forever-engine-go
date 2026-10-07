package elemental

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/shaman"
)

// Totem of Thunder (228176, patch 97): "Increases the critical strike chance of Lightning Bolt by 1%" (461295,
// a flat crit modifier on class mask word 0 bit 0). Every Lightning Bolt rank and Lightning Overload's
// Lightning Bolt rows carry that bit, so both crit 1% more with the relic; no other spell moves.
func TestTotemOfThunderAddsOnePercentLightningBoltCrit(t *testing.T) {
	const totemOfThunder = 228176

	mod := spelldata.MustFind(461295).EffectN(1)
	for _, id := range []int32{403, 15208, 408439, 408477} {
		if !mod.Covers(spelldata.MustFind(id)) {
			t.Fatalf("461295's mask does not name spell %d", id)
		}
	}
	for _, id := range []int32{421, 10605, 408479, 408484} {
		if mod.Covers(spelldata.MustFind(id)) {
			t.Fatalf("461295's mask names Chain Lightning row %d", id)
		}
	}

	type row struct {
		mask int64
		crit float64
	}
	// Each spell's bonus crit chance in percent, keyed by its id and whether it is an overload, since an
	// overload is registered on its rank's id.
	critBySpell := func(relic int32) map[string]row {
		items := make([]*proto.ItemSpec, 17)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: relic}
		player := core.WithSpec(&proto.Player{
			Race: proto.Race_RaceTroll, Class: proto.Class_ClassShaman, Equipment: &proto.EquipmentSpec{Items: items},
			Consumables: &proto.ConsumesSpec{}, TalentsString: DefaultTalents,
			Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}, &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}}}})
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 1},
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
		}, simsignals.CreateSignals())
		sim.Reset()
		character := sim.Raid.Parties[0].Players[0].GetCharacter()
		if got := character.Equipment[proto.ItemSlot_ItemSlotRanged].ID; got != relic {
			t.Fatalf("relic slot holds %d, want %d", got, relic)
		}
		out := map[string]row{}
		for _, spell := range character.Spellbook {
			key := spell.ActionID.String()
			if spell.ClassSpellMask&shaman.SpellMaskOverload != 0 {
				key += " overload"
			}
			out[key] = row{spell.ClassSpellMask, spell.BonusCritPercent}
		}
		return out
	}

	before, after := critBySpell(0), critBySpell(totemOfThunder)
	if len(before) != len(after) {
		t.Fatalf("%d spells without the relic, %d with it", len(before), len(after))
	}
	bolts, overloads := 0, 0
	for key, spell := range after {
		gained := spell.crit - before[key].crit
		switch {
		case spell.mask&shaman.SpellMaskLightningBolt != 0:
			bolts++
			if !near(gained, 1) {
				t.Errorf("Lightning Bolt %s: crit chance gains %v, want 1", key, gained)
			}
		case spell.mask&shaman.SpellMaskLightningBoltOverload != 0:
			overloads++
			if !near(gained, 1) {
				t.Errorf("Lightning Bolt overload %s: crit chance gains %v, want 1", key, gained)
			}
		default:
			if !near(gained, 0) {
				t.Errorf("%s (mask %b): crit chance changes by %v, want unchanged", key, spell.mask, gained)
			}
		}
	}
	if bolts != 10 || overloads != 10 {
		t.Fatalf("compared %d Lightning Bolt ranks and %d overloads, want 10 of each", bolts, overloads)
	}
}

func near(a, b float64) bool {
	d := a - b
	return d < 1e-9 && d > -1e-9
}
