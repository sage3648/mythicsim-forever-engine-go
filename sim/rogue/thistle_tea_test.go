package rogue

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Thistle Tea is used on its own once the bar has room for all but core.ThistleTeaSpill of its 100
// energy. With 10, a 100-energy Combat Rogue reached that only in 0.3% of fights, since energy refills
// at 10 a second and a bar is rarely under 10 when a global cooldown ends, while a Gnome's 105 bar
// reached it in every fight.
func TestThistleTeaIsUsedAtTheSpillThreshold(t *testing.T) {
	for _, tc := range []struct {
		race      proto.Race
		maxEnergy float64
	}{
		{proto.Race_RaceHuman, 100},
		{proto.Race_RaceGnome, 105},
	} {
		player := core.WithSpec(&proto.Player{
			Race: tc.race, Class: proto.Class_ClassRogue, Equipment: daggersOnly(),
			Consumables:   &proto.ConsumesSpec{ConjuredId: 7676, ConjuredItems: []int32{7676}},
			TalentsString: CombatTalents, Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}, DefaultOptions)
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 1},
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
		}, simsignals.CreateSignals())
		sim.Reset()
		rogue := sim.Raid.Parties[0].Players[0].(RogueAgent).GetRogue()
		if got := rogue.MaximumEnergy(); got != tc.maxEnergy {
			t.Fatalf("%v: maximum energy %v, want %v", tc.race, got, tc.maxEnergy)
		}
		tea := rogue.GetMajorCooldown(core.ActionID{ItemID: 7676})
		if tea == nil {
			t.Fatalf("%v: no Thistle Tea cooldown", tc.race)
		}
		metrics := rogue.NewEnergyMetrics(core.ActionID{OtherID: 1})
		threshold := tc.maxEnergy - 100 + core.ThistleTeaSpill
		for _, energy := range []float64{threshold, threshold + 1} {
			rogue.SpendEnergy(sim, rogue.CurrentEnergy(), metrics)
			rogue.AddEnergy(sim, energy, metrics)
			want := energy <= threshold
			if got := tea.ShouldActivate(sim, rogue.GetCharacter()); got != want {
				t.Errorf("%v at %v of %v energy: drinks = %v, want %v", tc.race, energy, tc.maxEnergy, got, want)
			}
		}
	}
}
