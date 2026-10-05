package dps

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func dragonsCallRaid() *proto.Raid {
	player := &proto.Player{
		Name: "warrior", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc, TalentsString: FuryTalents,
		Equipment: weaponsOnly(10847, 0),
		Spec:      DefaultOptions,
		Rotation:  core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "dps_reck").Rotation,
	}
	return &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}, NumActiveParties: 1}
}

// Dragon's Call's whelp is a gear pet: it has to be built at construction, get summoned by the
// proc, and then melee and spit Acid Spit.
func TestDragonsCallWhelp(t *testing.T) {
	res := core.RunRaidSim(&proto.RaidSimRequest{Raid: dragonsCallRaid(), Encounter: core.MakeSingleTargetEncounter(0), SimOptions: &proto.SimOptions{Iterations: 20, RandomSeed: 1}})
	if res.Error != nil {
		t.Fatal(res.Error.Message)
	}

	pets := res.RaidMetrics.Parties[0].Players[0].Pets
	if len(pets) != 1 || pets[0].Dps.Avg <= 0 {
		t.Fatalf("expected one whelp dealing damage, got %v", pets)
	}
	for _, action := range pets[0].Actions {
		if action.Id.GetSpellId() == 9591 {
			return
		}
	}
	t.Fatal("whelp never cast Acid Spit")
}

// Dragon's Call (13049) has a 45 sec category cooldown in the Forever client (60 sec on Classic
// Era). It gates the proc, so each whelp lives its full 15 sec and is never refreshed, and the next
// one comes at least 45 sec after the last. Without it, 1 PPM refreshed the whelp back to back.
func TestDragonsCallWhelpProcCooldown(t *testing.T) {
	logLine := regexp.MustCompile(`^\[(\d+\.\d+)\] \[[^\]]*Emerald Dragon Whelp\] Pet (summoned|dismissed)`)
	const summonGap, lifetime = 45 * time.Second, 15 * time.Second
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 600

	summons := 0
	for seed := int64(1); seed <= 5; seed++ {
		res := core.RunRaidSim(&proto.RaidSimRequest{Raid: dragonsCallRaid(), Encounter: encounter, SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, DebugFirstIteration: true}})
		if res.Error != nil {
			t.Fatal(res.Error.Message)
		}
		lastSummon := time.Duration(-1)
		for _, line := range strings.Split(res.Logs, "\n") {
			m := logLine.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			secs, _ := strconv.ParseFloat(m[1], 64)
			at := time.Duration(secs * float64(time.Second))
			switch m[2] {
			case "summoned":
				if lastSummon >= 0 && at-lastSummon < summonGap-10*time.Millisecond {
					t.Fatalf("seed %d: whelp summoned at %s, only %s after the last", seed, at, at-lastSummon)
				}
				lastSummon = at
				summons++
			case "dismissed":
				if lastSummon < 0 {
					continue // the reset dismissal at 0
				}
				lived := at - lastSummon
				cutShort := at >= time.Duration(encounter.Duration)*time.Second-10*time.Millisecond
				if lived > lifetime+10*time.Millisecond || (lived < lifetime-10*time.Millisecond && !cutShort) {
					t.Fatalf("seed %d: whelp summoned at %s lived %s, want 15s", seed, lastSummon, lived)
				}
			}
		}
	}
	if summons < 10 {
		t.Fatalf("only %d whelps over five 10 minute fights", summons)
	}
}
