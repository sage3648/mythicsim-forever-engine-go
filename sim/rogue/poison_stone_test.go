package rogue

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func newPoisonedRogue(t *testing.T, mhStone, ohStone int32) *Rogue {
	t.Helper()
	player := core.WithSpec(&proto.Player{
		Race:      proto.Race_RaceHuman,
		Class:     proto.Class_ClassRogue,
		Equipment: daggersOnly(),
		Consumables: &proto.ConsumesSpec{MhImbueId: instantImbueID, OhImbueId: deadlyImbueID,
			MhTempEnchantId: mhStone, OhTempEnchantId: ohStone},
		TalentsString: SubtletyTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, DefaultOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].(RogueAgent).GetRogue()
}

// Forever lets a Rogue's poisons sit beside a stone on the same weapon (Blizzard's Rogue and Warlock
// deep dive, 8 October 2026): the stone goes in the temporary enchantment slot (mythicsim patch 100)
// and both hands keep their poison.
func TestPoisonsStackWithStones(t *testing.T) {
	plain := newPoisonedRogue(t, 0, 0)
	stoned := newPoisonedRogue(t, 22756, 16138) // Elemental Sharpening Stone, Dense Sharpening Stone

	if got := stoned.GetStat(stats.PhysicalCritPercent) - plain.GetStat(stats.PhysicalCritPercent); got < 1.999 || got > 2.001 {
		t.Errorf("Elemental Sharpening Stone beside Instant Poison adds %v%% crit, want 2", got)
	}
	if got := stoned.AutoAttacks.OH().BaseDamageMax - plain.AutoAttacks.OH().BaseDamageMax; got != 8 {
		t.Errorf("Dense Sharpening Stone beside Deadly Poison adds %v off-hand damage, want 8", got)
	}
	if stoned.getPoisonProcMask(instantImbueID) != core.ProcMaskMeleeMH || stoned.getPoisonProcMask(deadlyImbueID) != core.ProcMaskMeleeOH {
		t.Errorf("poisons moved with the stones: instant %v, deadly %v", stoned.getPoisonProcMask(instantImbueID), stoned.getPoisonProcMask(deadlyImbueID))
	}
}
