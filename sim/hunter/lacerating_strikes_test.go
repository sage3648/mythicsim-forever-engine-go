package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Savage Strikes (19159) and Predator's Edge (1310627) name the Lacerating Strikes bleed (1310536) in
// their class masks, so its ticks get the same crit chance and crit damage as Mongoose Bite.
func TestLaceratingStrikesGetsSavageStrikesAndPredatorsEdge(t *testing.T) {
	player := &proto.Player{
		Name: "sv", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, TalentsString: SurvivalTalents,
		Equipment: WeaponsOnly,
		Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
			Ammo: proto.HunterOptions_Doomshot, PetType: proto.HunterOptions_Cat}}}},
	}
	raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}, NumActiveParties: 1}
	env, _, _ := core.NewEnvironment(raid, core.MakeSingleTargetEncounter(0), false, true)
	hunter := env.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()

	bite, bleed := hunter.MongooseBite, hunter.LaceratingStrikes
	if bleed.BonusCritPercent == 0 || bleed.BonusCritPercent != bite.BonusCritPercent {
		t.Errorf("bleed bonus crit %v%%, Mongoose Bite %v%%", bleed.BonusCritPercent, bite.BonusCritPercent)
	}
	if bleed.CritMultiplierAdditive == 0 || bleed.CritMultiplierAdditive != bite.CritMultiplierAdditive {
		t.Errorf("bleed crit damage bonus %v, Mongoose Bite %v", bleed.CritMultiplierAdditive, bite.CritMultiplierAdditive)
	}
}

// A Mongoose Bite that lands while the last bleed still runs replaces it with its own 40%. The bleed
// used to be written before the cast, and the refresh's expiry zeroed it, so every bleed after the
// first in a fight ticked for nothing.
func TestLaceratingStrikesRefreshKeepsTheNewBleed(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name: "sv", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc, TalentsString: SurvivalMeleeTalents,
			Equipment: WeaponsOnly,
			Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
				Ammo: proto.HunterOptions_Doomshot, PetType: proto.HunterOptions_PetNone}}}},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
	if hunter.LaceratingStrikes == nil {
		t.Fatal("the melee Survival build no longer takes Lacerating Strikes")
	}
	target := hunter.CurrentTarget
	dot := hunter.LaceratingStrikes.Dot(target)

	for i, bite := range []float64{1000, 1400} {
		hunter.procLaceratingStrikes(sim, &core.SpellResult{Target: target, Damage: bite})
		if !dot.IsActive() {
			t.Fatalf("bite %d: no bleed", i+1)
		}
		want := bite * 0.4 / float64(dot.BaseTickCount)
		if dot.SnapshotBaseDamage != want || dot.SnapshotAttackerMultiplier != 1 {
			t.Errorf("bite %d: bleed ticks %v at x%v, want %v at x1", i+1, dot.SnapshotBaseDamage, dot.SnapshotAttackerMultiplier, want)
		}
	}
}
