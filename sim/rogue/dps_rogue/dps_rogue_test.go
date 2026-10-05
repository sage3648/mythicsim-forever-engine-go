package dpsrogue

import (
	"github.com/wowsims/classic/sim/arenalib"
	"testing"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"github.com/wowsims/classic/sim/core/simsignals"
)

func init() {
	RegisterDpsRogue()
}

func TestCombatSinisterStrike(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassRogue,
			Race:       proto.Race_RaceHuman,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},

			Talents:     CombatSwordsTalents,
			GearSet:     core.GetGearSet("../../../ui/rogue/gear_sets", "combat_sinister_strike_prebis"),
			Rotation:    core.GetAplRotation("../../../ui/rogue/apls", "combat_sinister_strike"),
			Buffs:       core.ForeverBuffs,
			Consumes:    Phase1Consumes,
			Phase:       1,
			SpecOptions: core.SpecOptionsCombo{Label: "Windfury MH", SpecOptions: DefaultRogue},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh:    Stats,

			Ruleset: proto.Ruleset_RulesetForever,
		},
	}))
}

func TestCombatDaggers(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassRogue,
			Race:       proto.Race_RaceHuman,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},

			Talents:     CombatDaggersTalents,
			GearSet:     core.GetGearSet("../../../ui/rogue/gear_sets", "combat_backstab_prebis"),
			Rotation:    core.GetAplRotation("../../../ui/rogue/apls", "combat_backstab"),
			Buffs:       core.ForeverBuffs,
			Consumes:    Phase1Consumes,
			Phase:       1,
			SpecOptions: core.SpecOptionsCombo{Label: "Windfury MH", SpecOptions: DefaultRogue},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh:    Stats,

			Ruleset: proto.Ruleset_RulesetForever,
		},
	}))
}

func TestAssassinationMutilate(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassRogue,
			Race:       proto.Race_RaceHuman,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},

			Talents:     AssassinationMutilateTalents,
			GearSet:     core.GetGearSet("../../../ui/rogue/gear_sets", "combat_backstab_prebis"),
			Rotation:    core.GetAplRotation("../../../ui/rogue/apls", "forever_mutilate"),
			Buffs:       core.ForeverBuffs,
			Consumes:    Phase1PoisonConsumes,
			Phase:       1,
			SpecOptions: core.SpecOptionsCombo{Label: "Poisons", SpecOptions: DefaultRogue},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh:    Stats,

			Ruleset: proto.Ruleset_RulesetForever,
		},
	}))
}

func TestColdBloodAffectsBothMutilateStrikes(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Ruleset: proto.Ruleset_RulesetForever},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Class:         proto.Class_ClassRogue,
			Race:          proto.Race_RaceHuman,
			Equipment:     core.GetGearSet("../../../ui/rogue/gear_sets", "combat_backstab_prebis").GearSet,
			TalentsString: AssassinationMutilateTalents,
			Rotation:      core.GetAplRotation("../../../ui/rogue/apls", "forever_mutilate").Rotation,
			Spec:          DefaultRogue,
		}, nil, nil, nil),
		Encounter: core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()

	rogue := sim.Raid.Parties[0].Players[0].(*DpsRogue).Rogue
	mainHand := rogue.Mutilate
	if mainHand == nil {
		t.Fatal("Mutilate is not registered")
	}
	offHand := rogue.GetSpell(mainHand.ActionID.WithTag(2))
	coldBloodAura := rogue.GetAura("Cold Blood")
	if offHand == nil || rogue.ColdBlood == nil || coldBloodAura == nil {
		t.Fatal("Mutilate or Cold Blood is not registered")
	}

	mainHandBaseCrit := mainHand.BonusCritRating
	offHandBaseCrit := offHand.BonusCritRating
	rogue.ColdBlood.Cast(sim, rogue.CurrentTarget)
	bonus := float64(100 * core.CritRatingPerCritChance)
	if mainHand.BonusCritRating != mainHandBaseCrit+bonus || offHand.BonusCritRating != offHandBaseCrit+bonus {
		t.Fatal("Cold Blood did not apply its crit bonus to both Mutilate strikes")
	}

	coldBloodAura.OnSpellHitDealt(coldBloodAura, sim, mainHand, &core.SpellResult{})
	if !coldBloodAura.IsActive() || offHand.BonusCritRating != offHandBaseCrit+bonus {
		t.Fatal("Cold Blood was consumed before Mutilate's off-hand strike")
	}

	coldBloodAura.OnSpellHitDealt(coldBloodAura, sim, offHand, &core.SpellResult{})
	if coldBloodAura.IsActive() || mainHand.BonusCritRating != mainHandBaseCrit || offHand.BonusCritRating != offHandBaseCrit {
		t.Fatal("Cold Blood was not consumed after both Mutilate strikes")
	}
}

func TestSubtletyHemorrhage(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassRogue,
			Race:       proto.Race_RaceHuman,
			OtherRaces: []proto.Race{proto.Race_RaceOrc},

			Talents:     SubtletyHemorrhageTalents,
			GearSet:     core.GetGearSet("../../../ui/rogue/gear_sets", "combat_backstab_prebis"),
			Rotation:    core.GetAplRotation("../../../ui/rogue/apls", "forever_hemorrhage"),
			Buffs:       core.ForeverBuffs,
			Consumes:    Phase1PoisonConsumes,
			Phase:       1,
			SpecOptions: core.SpecOptionsCombo{Label: "Poisons", SpecOptions: DefaultRogue},

			ItemFilter:      ItemFilters,
			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh:    Stats,

			Ruleset: proto.Ruleset_RulesetForever,
		},
	}))
}

var CombatSwordsTalents = "00530310501-32003311201515231"
var CombatDaggersTalents = "005302005-30230320201515231-102"
var AssassinationMutilateTalents = "00530310551021051-302303202004"
var SubtletyHemorrhageTalents = "125320101--5320003310013211551"

var DefaultRogue = &proto.Player_Rogue{
	Rogue: &proto.Rogue{
		Options: &proto.RogueOptions{},
	},
}

var ItemFilters = core.ItemFilter{
	ArmorType: proto.ArmorType_ArmorTypeLeather,
	WeaponTypes: []proto.WeaponType{
		proto.WeaponType_WeaponTypeDagger,
		proto.WeaponType_WeaponTypeFist,
		proto.WeaponType_WeaponTypeSword,
		proto.WeaponType_WeaponTypeMace,
	},
	RangedWeaponTypes: []proto.RangedWeaponType{
		proto.RangedWeaponType_RangedWeaponTypeBow,
		proto.RangedWeaponType_RangedWeaponTypeCrossbow,
		proto.RangedWeaponType_RangedWeaponTypeGun,
	},
}

var Stats = []proto.Stat{
	proto.Stat_StatAttackPower,
	proto.Stat_StatAgility,
	proto.Stat_StatStrength,
	proto.Stat_StatMeleeHit,
	proto.Stat_StatMeleeCrit,
}

// Combat runs Windfury on the main hand and a poison off hand; the Assassination combo
// below is the one that poisons both.
var Phase1Consumes = core.ConsumesCombo{
	Label: "P1-Consumes",
	Consumes: &proto.Consumes{
		AgilityElixir:   proto.AgilityElixir_ElixirOfTheMongoose,
		MainHandImbue:   proto.WeaponImbue_Windfury,
		OffHandImbue:    proto.WeaponImbue_InstantPoison,
		StrengthBuff:    proto.StrengthBuff_JujuPower,
		AttackPowerBuff: proto.AttackPowerBuff_JujuMight,
	},
}

// Venom and Mutilate both key off the poisons, so the Assassination build runs them on both weapons.
var Phase1PoisonConsumes = core.ConsumesCombo{
	Label: "P1-Poison-Consumes",
	Consumes: &proto.Consumes{
		AgilityElixir:   proto.AgilityElixir_ElixirOfTheMongoose,
		MainHandImbue:   proto.WeaponImbue_InstantPoison,
		OffHandImbue:    proto.WeaponImbue_DeadlyPoison,
		StrengthBuff:    proto.StrengthBuff_JujuPower,
		AttackPowerBuff: proto.AttackPowerBuff_JujuMight,
	},
}

// The arena entry for this spec. Skipped unless ARENA_OUT is set; see sim/arenalib.
func TestArena(t *testing.T) {
	arenalib.Run(t, arenalib.Spec{
		Dir:         "rogue",
		Class:       proto.Class_ClassRogue,
		Race:        proto.Race_RaceHuman,
		SpecOptions: DefaultRogue,
		Role:        arenalib.Melee,
		// Poisons are a rogue ability, not something on the vendor list.
		ClassImbues: arenalib.ClassImbues{OffHand: proto.WeaponImbue_InstantPoison},
		Buffs:       core.ForeverBuffs,
	})
}
