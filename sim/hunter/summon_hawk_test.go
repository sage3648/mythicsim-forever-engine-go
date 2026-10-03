package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Summon Hawk keeps two hawks out at once (1293527's third effect), each assaulting on its own, and
// their hits can crit so Ferocity reaches them. The dive bomb always hits (client always-hit attribute).
func TestSummonHawkTwoHawks(t *testing.T) {
	for _, front := range []bool{false, true} {
		t.Run(map[bool]string{false: "behind", true: "in_front"}[front], func(t *testing.T) {
			player := &proto.Player{
				Name: "bm", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, TalentsString: BeastMasteryTalents,
				Equipment: WeaponsOnly, DistanceFromTarget: 30, InFrontOfTarget: front,
				Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
					Ammo: proto.HunterOptions_Doomshot, QuiverBonus: proto.HunterOptions_Speed15, PetType: proto.HunterOptions_Cat,
					PetAttackSpeed: proto.HunterOptions_OneTwo, PetUptime: 1}}}},
				Rotation: core.GetAplRotation("../../ui/specs/hunter/dps/apls", "bm").Rotation,
			}
			raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}, NumActiveParties: 1}
			res := core.RunRaidSim(&proto.RaidSimRequest{Raid: raid, Encounter: core.MakeSingleTargetEncounter(0), SimOptions: &proto.SimOptions{Iterations: 10, RandomSeed: 1}})
			if res.Error != nil {
				t.Fatal(res.Error.Message)
			}

			hits := map[int32]int32{}
			var critHits, diveBombMisses, misses, dodges, parries int32
			for _, action := range res.RaidMetrics.Parties[0].Players[0].Actions {
				if action.Id.GetSpellId() != spellData.SummonHawk.Highest().ID {
					continue
				}
				for _, target := range action.Targets {
					if action.Id.GetTag() == 0 {
						diveBombMisses += target.Misses + target.Dodges + target.Parries + target.Blocks
					}
					if action.Id.GetTag() != 0 {
						hits[action.Id.GetTag()] += target.Hits + target.Crits
						critHits += target.Crits
						misses += target.Misses
						dodges += target.Dodges
						parries += target.Parries
						if target.Ticks+target.CritTicks != 0 {
							t.Fatal("Hawk attacks counted as unavoidable periodic ticks")
						}
					}
				}
			}
			if hits[1] == 0 || hits[2] == 0 || hits[3] != 0 {
				t.Fatalf("hawk hits by hawk: %v, want two hawks both assaulting", hits)
			}
			if diveBombMisses != 0 {
				t.Fatalf("%d dive bombs missed, were dodged, parried or blocked; the client says they always hit", diveBombMisses)
			}
			if critHits == 0 {
				t.Fatal("no hawk hit crit")
			}
			if misses == 0 || dodges == 0 {
				t.Fatalf("Hawk avoidance: %d misses, %d dodges; want both recorded", misses, dodges)
			}
			if front && parries == 0 {
				t.Fatal("Hawk attacks in front of the target never parried")
			}
			if !front && parries != 0 {
				t.Fatal("Hawk attacks behind the target parried")
			}
		})
	}
}
