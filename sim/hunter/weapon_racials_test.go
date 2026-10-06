package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

type hunterCritChances struct {
	autoShot, aimedShot, multiShot, melee float64
}

func weaponRacialCritChances(t *testing.T, race proto.Race, mainHand int32, disableWeaponSpec bool) hunterCritChances {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name: "mm", Class: proto.Class_ClassHunter, Race: race, TalentsString: MarksmanshipTalents,
			Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
				{Id: mainHand},
				{},
				{Id: 18713}, // Rhok'delar, Longbow of the Ancient Keepers
			}},
			DisableWeaponSpecialization: disableWeaponSpec,
			Rotation:                    &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
				Ammo: proto.HunterOptions_Doomshot, PetType: proto.HunterOptions_PetNone}}}},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
	at := hunter.AttackTables[hunter.CurrentTarget.UnitIndex]
	if hunter.AimedShot == nil || hunter.MultiShot == nil {
		t.Fatal("the Marksmanship build no longer has Aimed Shot and Multi-Shot")
	}
	return hunterCritChances{
		autoShot:  hunter.AutoAttacks.RangedAuto().PhysicalCritChance(at),
		aimedShot: hunter.AimedShot.PhysicalCritChance(at),
		multiShot: hunter.MultiShot.PhysicalCritChance(at),
		melee:     hunter.AutoAttacks.MHAuto().PhysicalCritChance(at),
	}
}

// Human Sword, Orc Axe and Dwarf Mace Specialization raise melee and ability crit but leave Auto
// Shot alone (forever-bugs #91, patch 88). Each race is compared with itself with its weapon racial
// switched off, so base agility drops out; a Night Elf, who has no weapon racial, moves nowhere.
func TestWeaponRacialsSkipTheRangedAutoAttack(t *testing.T) {
	const (
		arcaniteChampion = 12790 // two-hand sword
		arcaniteReaper   = 12784 // two-hand axe
		ironfoe          = 11684 // one-hand mace
	)
	for _, tc := range []struct {
		race     proto.Race
		mainHand int32
		bonus    float64
	}{
		{proto.Race_RaceHuman, arcaniteChampion, 0.02},
		{proto.Race_RaceOrc, arcaniteReaper, 0.01},
		{proto.Race_RaceDwarf, ironfoe, 0.01},
		{proto.Race_RaceNightElf, arcaniteChampion, 0},
	} {
		t.Run(tc.race.String(), func(t *testing.T) {
			with := weaponRacialCritChances(t, tc.race, tc.mainHand, false)
			without := weaponRacialCritChances(t, tc.race, tc.mainHand, true)
			for _, c := range []struct {
				name string
				got  float64
				want float64
			}{
				{"Auto Shot", with.autoShot - without.autoShot, 0},
				{"Aimed Shot", with.aimedShot - without.aimedShot, tc.bonus},
				{"Multi-Shot", with.multiShot - without.multiShot, tc.bonus},
				{"melee auto attack", with.melee - without.melee, tc.bonus},
			} {
				if !core.WithinToleranceFloat64(c.want, c.got, 1e-9) {
					t.Errorf("%s crit chance moves %.4f with the weapon racial, want %.4f", c.name, c.got, c.want)
				}
			}
		})
	}
}
