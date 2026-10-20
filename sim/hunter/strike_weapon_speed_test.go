package hunter

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Strider Kick and Mongoose Bite take the higher of the main hand's real and normalized speed (patch 187): a 3.8
// two-hander keeps 3.8, a 1.7 one-hander is lifted to its normalized 2.4, and with the switch off both take the
// normalized speed.
func TestStrikesKeepSlowWeaponSpeed(t *testing.T) {
	player := &proto.Player{
		Name: "strike speed", Class: proto.Class_ClassHunter, Race: proto.Race_RaceNightElf, TalentsString: SurvivalTalents,
		Equipment: WeaponsOnly, Consumables: &proto.ConsumesSpec{},
		Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{
			Ammo: proto.HunterOptions_Doomshot, PetType: proto.HunterOptions_PetNone}}}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
	}, simsignals.CreateSignals())
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(HunterAgent).GetHunter()
	mh := hunter.AutoAttacks.MH()
	saved := *mh
	defer func() { *mh = saved; StrikesKeepSlowWeaponSpeed = true }()

	// a fixed base damage of 100 and 1,400 attack power: 100 + speed x 100
	mh.BaseDamageMin, mh.BaseDamageMax, mh.AttackPowerPerDPS = 100, 100, core.DefaultAttackPowerPerDPS
	for _, tc := range []struct {
		name             string
		real, normalized float64
		keepSlow         bool
		want             float64
	}{
		{"3.8 two-hander", 3.8, 3.3, true, 480},
		{"1.7 one-hander", 1.7, 2.4, true, 340},
		{"3.8 two-hander, switch off", 3.8, 3.3, false, 430},
	} {
		mh.SwingSpeed, mh.NormalizedSwingSpeed = tc.real, tc.normalized
		StrikesKeepSlowWeaponSpeed = tc.keepSlow
		if got := hunter.strikeWeaponDamage(sim, 1400); math.Abs(got-tc.want) > 1e-6 {
			t.Errorf("%s: strike weapon damage %.1f, want %.1f", tc.name, got, tc.want)
		}
	}
}
