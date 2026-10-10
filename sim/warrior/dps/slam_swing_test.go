package dps

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Slam is a 1.5 sec hard cast without Improved Slam. Client 1.60.1.70170 and docs/forever_rules.md
// say Slam no longer resets the swing timer, and a hard cast otherwise holds the swing, so a Slam
// moves neither a swing due during the cast nor one due after it.
func TestSlamLeavesTheSwingAlone(t *testing.T) {
	const improvedSlamUntaken = "32305213132515001-5502" // see TestImprovedSlamShortensSlamCooldown
	slamID := core.ActionID{SpellID: 11605}

	for _, tc := range []struct {
		name string
		// When the swing is due, counted from now.
		due func(speed time.Duration) time.Duration
	}{
		{"swing due during the cast", func(speed time.Duration) time.Duration { return 500*time.Millisecond - speed }},
		{"swing due after the cast", func(speed time.Duration) time.Duration { return 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid: core.SinglePlayerRaidProto(&proto.Player{
					Race: proto.Race_RaceHuman, Class: proto.Class_ClassWarrior,
					Equipment: TwoHandGear.GearSet, TalentsString: improvedSlamUntaken, Spec: DefaultOptions,
					Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  core.MakeSingleTargetEncounter(0),
				SimOptions: &proto.SimOptions{RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()
			war := sim.Raid.Parties[0].Players[0].(*DpsWarrior)
			war.AutoAttacks.EnableAutoSwing(sim)

			speed := war.AutoAttacks.MainhandSwingSpeed()
			war.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+tc.due(speed))
			before := war.AutoAttacks.MainhandSwingAt()

			war.AddRage(sim, 50, war.NewRageMetrics(slamID))
			if !war.GetSpell(slamID).Cast(sim, war.CurrentTarget) {
				t.Fatal("Slam did not cast")
			}
			if got := war.AutoAttacks.MainhandSwingAt(); got != before {
				t.Errorf("Slam moved the next swing from %v to %v", before, got)
			}
		})
	}
}
