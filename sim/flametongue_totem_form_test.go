package sim

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
	"google.golang.org/protobuf/encoding/protojson"
)

type flametongueTotemProbe struct {
	total, largestHit, want float64
}

// Casts the party totem's hit (16389) 3000 times under a druid in form wielding Soulkeeper, a 3.8 s
// staff, and records its damage: the total, and the largest hit that was neither a crit nor partly
// resisted (the log rounds it to 0.001).
func probeFlametongueTotem(t *testing.T, playerJSON string, form druid.DruidForm, paw, spellDamage float64) flametongueTotemProbe {
	t.Helper()
	player := &proto.Player{}
	if err := protojson.Unmarshal([]byte(playerJSON), player); err != nil {
		t.Fatal(err)
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 1607} // Soulkeeper, 3.8 s
	player.Equipment = &proto.EquipmentSpec{Items: items}
	player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{FlametongueTotem: true}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	d := sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	if !d.InForm(form) || d.AutoAttacks.MH().SwingSpeed != paw {
		t.Fatalf("want form %v with a %v s paw, got a %v s paw", form, paw, d.AutoAttacks.MH().SwingSpeed)
	}
	d.AddStatDynamic(sim, stats.SpellDamage, spellDamage)

	attack := d.GetSpell(core.ActionID{SpellID: 16389})
	if attack == nil {
		t.Fatal("the party totem's hit (16389) is not registered")
	}
	target := d.CurrentTarget
	var probe flametongueTotemProbe
	// Only a plain hit logs exactly "Hit for"; crits and partial resists name themselves.
	plainHit := regexp.MustCompile(`SpellID: 16389\S* Hit for ([0-9.]+) damage`)
	sim.Log = func(format string, args ...any) {
		if m := plainHit.FindStringSubmatch(fmt.Sprintf(format, args...)); m != nil {
			hit, _ := strconv.ParseFloat(m[1], 64)
			probe.largestHit = max(probe.largestHit, hit)
		}
	}
	for range 3000 {
		attack.Cast(sim, target)
	}
	probe.total = attack.SpellMetrics[target.UnitIndex].TotalDamage
	// 3.8 s of speed at 13.63 a second (16389's 1363), through the druid's own damage multipliers.
	probe.want = 3.8 * 13.63 * attack.AttackerDamageMultiplier(d.AttackTables[target.UnitIndex], false)
	return probe
}

// Hameru on the beta (MythicSim Discord, 6 October 2026): in Cat Form the party's Flametongue Totem adds
// the hit of the weapon in the main hand, by that weapon's speed, not by the 1.0 s paw's; and the hit
// takes none of the druid's spell power (patch 91). Bear Form is taken to follow the same rule.
func TestFlametongueTotemInFormUsesTheEquippedWeaponAndNoSpellPower(t *testing.T) {
	for _, row := range []struct {
		name   string
		player string
		form   druid.DruidForm
		paw    float64
	}{{"Cat", catJSON, druid.Cat, 1}, {"Bear", bearJSON, druid.Bear, 2.5}} {
		t.Run(row.name, func(t *testing.T) {
			plain := probeFlametongueTotem(t, row.player, row.form, row.paw, 0)
			if math.Abs(plain.largestHit-plain.want) > 0.001 {
				t.Errorf("largest plain hit %.4f, want %.4f (a 3.8 s weapon)", plain.largestHit, plain.want)
			}
			// The same seed rolls the same outcomes, so a coefficient would show in the total.
			if withPower := probeFlametongueTotem(t, row.player, row.form, row.paw, 1000); withPower.total != plain.total {
				t.Errorf("1000 spell power moves the totem's damage from %.2f to %.2f", plain.total, withPower.total)
			}
		})
	}
}
