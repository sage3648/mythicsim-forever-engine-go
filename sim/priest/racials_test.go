package priest

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func racialPriest(race proto.Race, disableRacials bool) *core.Character {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name:           "Priest",
			Race:           race,
			Class:          proto.Class_ClassPriest,
			Equipment:      &proto.EquipmentSpec{},
			Consumables:    &proto.ConsumesSpec{},
			TalentsString:  ShadowTalents,
			Rotation:       &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			Spec:           &proto.Player_DpsPriest{DpsPriest: &proto.DpsPriest{Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}}}},
			DisableRacials: disableRacials,
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Starshards is the night elf priest's race ability, so disabling racials takes it away.
func TestDisableRacialsDropsStarshards(t *testing.T) {
	starshards := core.ActionID{SpellID: 19305} // rank 7
	if racialPriest(proto.Race_RaceNightElf, false).GetSpell(starshards) == nil {
		t.Fatal("a night elf priest has no Starshards, so this test checks nothing")
	}
	if racialPriest(proto.Race_RaceNightElf, true).GetSpell(starshards) != nil {
		t.Error("a night elf priest with racials disabled still has Starshards")
	}
}

// Dark Sacrifice is the undead priest's race ability (SkillLineAbility race mask 16), so a priest
// of any other race, or an undead one with racials disabled, has neither the spell nor the major
// cooldown that Autocast Other Cooldowns would fire.
func TestDarkSacrificeIsUndeadOnly(t *testing.T) {
	darkSacrifice := core.ActionID{SpellID: 1277328} // rank 5
	for _, c := range []struct {
		race           proto.Race
		disableRacials bool
		want           bool
	}{
		{proto.Race_RaceUndead, false, true},
		{proto.Race_RaceUndead, true, false},
		{proto.Race_RaceHuman, false, false},
		{proto.Race_RaceNightElf, false, false},
		{proto.Race_RaceTroll, false, false},
		{proto.Race_RaceDwarf, false, false},
	} {
		character := racialPriest(c.race, c.disableRacials)
		if got := character.GetSpell(darkSacrifice) != nil; got != c.want {
			t.Errorf("%s (racials disabled %v): has Dark Sacrifice = %v, want %v", c.race, c.disableRacials, got, c.want)
		}
		if got := character.GetMajorCooldown(darkSacrifice) != nil; got != c.want {
			t.Errorf("%s (racials disabled %v): Dark Sacrifice major cooldown = %v, want %v", c.race, c.disableRacials, got, c.want)
		}
	}
}
