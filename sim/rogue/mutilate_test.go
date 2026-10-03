package rogue

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Mutilate's cast rolls miss/dodge/parry once; its two hand strikes carry the client's No Attack
// Miss/Dodge/Parry attributes and never roll them again. Beta logs agree: in foreverlogs 2701/2702
// the level 30 cast (1310707) missed, was dodged or parried 16 times in 162, the hand strikes
// (1310705/1310706) 0 times in 280.
func TestMutilateHitsDoNotRollAvoidance(t *testing.T) {
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Race: proto.Race_RaceHuman, Class: proto.Class_ClassRogue,
			Equipment: DefaultGear.GearSet, TalentsString: AssassinationTalents, Spec: DefaultOptions,
			Rotation: core.GetAplRotation("../../ui/specs/rogue/dps/apls", "forever_mutilate").Rotation,
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 5, RandomSeed: 1},
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	casts, castAvoided, hits := int32(0), int32(0), int32(0)
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() != MutilateSpellID {
			continue
		}
		for _, target := range action.Targets {
			avoided := target.Misses + target.Dodges + target.Parries
			if action.Id.GetTag() == 0 {
				casts += target.Casts
				castAvoided += avoided
			} else {
				hits += target.Hits + target.Crits
				if avoided > 0 {
					t.Errorf("Mutilate hand strike (tag %d) was avoided %d times", action.Id.GetTag(), avoided)
				}
			}
		}
	}
	if casts == 0 || castAvoided == 0 || hits == 0 {
		t.Errorf("want Mutilate casts, some avoided and hand strikes: casts %d, avoided %d, strikes %d", casts, castAvoided, hits)
	}
}
