package sim

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

func runArmsWarrior(t *testing.T, seed int64) *proto.RaidSimResult {
	t.Helper()
	player := &proto.Player{}
	if err := protojson.Unmarshal([]byte(warriorJSON), player); err != nil {
		t.Fatal(err)
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 15273} // Death Striker, 2.8 s
	player.Equipment = &proto.EquipmentSpec{Items: items}
	player.Rotation = core.GetAplRotation("../ui/specs/warrior/dps/apls", "dps_no_reck").Rotation
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 20, RandomSeed: seed},
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	return result
}

func damageRanges(tam *proto.TargetedActionMetrics) map[string]*proto.DamageRange {
	return map[string]*proto.DamageRange{"hit": tam.HitRange, "crit": tam.CritRange, "tick": tam.TickRange, "crit tick": tam.CritTickRange}
}

// Each action reports the spread of its landed damage per target (patch 92): direct hits, direct crits,
// ticks and critical ticks, each with a count, total, smallest and largest. Together with the glancing,
// blocked and crushing totals they add up to the action's damage.
func TestActionMetricsCarryDamageRanges(t *testing.T) {
	result := runArmsWarrior(t, 1)
	seen := map[string]bool{}
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		for _, tam := range action.Targets {
			sum := tam.GlanceDamage + tam.BlockDamage + tam.BlockedCritDamage + tam.CrushDamage
			for kind, r := range damageRanges(tam) {
				if r == nil {
					continue
				}
				seen[kind] = true
				if r.Count <= 0 || r.Min > r.Total/float64(r.Count)+1e-9 || r.Max < r.Total/float64(r.Count)-1e-9 {
					t.Errorf("%v %s range %v is not a range", action.Id, kind, r)
				}
				sum += r.Total
			}
			if math.Abs(sum-tam.Damage) > 1e-6*max(1, tam.Damage) {
				t.Errorf("%v: ranges and the other kinds add up to %.3f, damage is %.3f", action.Id, sum, tam.Damage)
			}
			if tam.HitRange != nil && tam.CritRange != nil && tam.CritRange.Max < tam.HitRange.Max {
				t.Errorf("%v: largest crit %.1f below the largest hit %.1f", action.Id, tam.CritRange.Max, tam.HitRange.Max)
			}
		}
	}
	for _, kind := range []string{"hit", "crit", "tick"} {
		if !seen[kind] {
			t.Errorf("no action has a %s range; an Arms warrior with Deep Wounds should", kind)
		}
	}
}

// Concurrent sims combine the ranges: counts and totals add, the smallest and largest are the extremes.
func TestConcurrentResultsCombineDamageRanges(t *testing.T) {
	first, second := runArmsWarrior(t, 1), runArmsWarrior(t, 2)
	combined := core.CombineConcurrentSimResults([]*proto.RaidSimResult{first, second}, false)
	byID := func(r *proto.RaidSimResult) map[string]*proto.ActionMetrics {
		out := map[string]*proto.ActionMetrics{}
		for _, a := range r.RaidMetrics.Parties[0].Players[0].Actions {
			out[a.Id.String()] = a
		}
		return out
	}
	a, b := byID(first), byID(second)
	for id, action := range byID(combined) {
		for i, tam := range action.Targets {
			for kind, got := range damageRanges(tam) {
				var want core.DamageRange
				for _, part := range []map[string]*proto.ActionMetrics{a, b} {
					if pa := part[id]; pa != nil {
						if r := damageRanges(pa.Targets[i])[kind]; r != nil {
							want.Count += r.Count
							want.Total += r.Total
							if want.Count == r.Count || r.Min < want.Min {
								want.Min = r.Min
							}
							want.Max = max(want.Max, r.Max)
						}
					}
				}
				if want.Count == 0 {
					if got != nil {
						t.Errorf("%s %s: combined %v from nothing", id, kind, got)
					}
					continue
				}
				if got == nil || got.Count != want.Count || math.Abs(got.Total-want.Total) > 1e-6 || got.Min != want.Min || got.Max != want.Max {
					t.Errorf("%s %s: combined %v, want %+v", id, kind, got, want)
				}
			}
		}
	}
}
