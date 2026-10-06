package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// A weaver that steps into melee with its swing already due queues Raptor Strike on arrival, and the
// Raptor Strike takes that first swing. The swing used to land as a white hit the moment the hunter
// arrived, before the rotation ran, so the Raptor Strike waited out a full swing timer behind it.
func TestRaptorStrikeTakesTheSwingOnArrival(t *testing.T) {
	rotation := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(`{"type": "TypeAPL", "priorityList": [
		{"action": {"castSpell": {"spellId": {"spellId": 14266, "tag": 3}}}},
		{"action": {"move": {"rangeFromTarget": {"const": {"val": "5"}}}}}
	]}`), rotation); err != nil {
		t.Fatal(err)
	}
	player := &proto.Player{
		Name: "weaver", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, TalentsString: SurvivalTalents,
		Equipment: WeaponsOnly, Consumables: &proto.ConsumesSpec{}, DistanceFromTarget: 11, ReactionTimeMs: 100,
		Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
			Ammo: proto.HunterOptions_Doomshot, QuiverBonus: proto.HunterOptions_Speed15, PetType: proto.HunterOptions_PetNone}}}},
		Rotation: rotation,
	}
	encounter := core.MakeSingleTargetEncounter(0)
	// The hunter walks the 6 yards in 0.857 seconds, between two of the rotation's 100ms checks, and
	// Arcanite Reaper swings every 3.8 seconds, so in a 3 second fight only the arrival swing lands.
	encounter.Duration = 3
	encounter.DurationVariation = 0
	result := core.RunRaidSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}

	casts := func(matches func(id *proto.ActionID) bool) int32 {
		var n int32
		for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
			if matches(action.Id) {
				for _, target := range action.Targets {
					n += target.Casts
				}
			}
		}
		return n
	}
	raptor := casts(func(id *proto.ActionID) bool { return id.GetSpellId() == 14266 && id.Tag == 1 })
	white := casts(func(id *proto.ActionID) bool { return id.GetOtherId() == proto.OtherAction_OtherActionAttack })
	if raptor != 1 || white != 0 {
		t.Errorf("the arrival swing landed %d Raptor Strikes and %d white hits, want 1 and 0", raptor, white)
	}
}
