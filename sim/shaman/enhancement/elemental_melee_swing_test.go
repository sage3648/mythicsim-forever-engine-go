package enhancement

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// An Elemental build with melee buffs, run through the Enhancement spec: Searing Totem pulses every
// 2.43 sec, and Lightning Bolt chain-casts 2 sec at a time next to a main-hand swing every 2.7 sec.
// The main-hand swing is the "{OtherID: 3, Tag: 1}" hit. Searing Totem's pulse ("Attack", spell
// 10436) is a fire totem spell, not a melee swing, so it may land inside a cast; the swing may not.
func TestElementalMeleeHoldsTheMainHandThroughBolts(t *testing.T) {
	// The rotation of the build's request: Searing Totem (rank 6) while its dot is down, and Lightning
	// Bolt (rank 10) while mana lasts.
	rotation := core.APLRotationFromJsonString(`{"type":"TypeAPL",` +
		`"prepullActions":[{"action":{"castSpell":{"spellId":{"spellId":15208}}},"doAtValue":{"const":{"val":"-2s"}}}],` +
		`"priorityList":[` +
		`{"action":{"castSpell":{"spellId":{"spellId":10438}},"condition":{"and":{"vals":[` +
		`{"cmp":{"lhs":{"numberTargets":{}},"op":"OpEq","rhs":{"const":{"val":"1"}}}},` +
		`{"not":{"val":{"dotIsActive":{"spellId":{"spellId":10438}}}}}]}}}},` +
		`{"action":{"castSpell":{"spellId":{"spellId":15208}},"condition":{"cmp":{"lhs":{"currentMana":{}},` +
		`"op":"OpGe","rhs":{"math":{"lhs":{"remainingTime":{}},"op":"OpMul","rhs":{"const":{"val":"35"}}}}}}}}]}`)
	items := []*proto.ItemSpec{
		{Id: 22267}, {Id: 22403}, {Id: 18681}, {Id: 18350}, {Id: 18385}, {Id: 18497}, {Id: 12632},
		{Id: 11662}, {Id: 252486}, {Id: 18322}, {Id: 22433}, {Id: 12545}, {Id: 18534}, {Id: 228176},
	}
	player := &proto.Player{
		Name: "Elemental melee", Class: proto.Class_ClassShaman, Race: proto.Race_RaceTauren,
		TalentsString: "5500301320123051-052-05305",
		Equipment:     &proto.EquipmentSpec{Items: items},
		Buffs:         &proto.IndividualBuffs{}, Consumables: &proto.ConsumesSpec{},
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}},
		}},
		Rotation: rotation,
	}
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1179730229, DebugFirstIteration: true},
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}

	line := regexp.MustCompile(`^\[(-?[0-9.]+)\] \[[^\]]*\] (.*)$`)
	melee := regexp.MustCompile(`^\[Target 1\] \{OtherID: 3, Tag: 1\} (Hit|Crit|Miss|Dodge|Parry|Glance)`)
	type cast struct{ start, end float64 }
	var casts []cast
	var swings, pulses []float64
	castStart := map[string]float64{}
	for _, raw := range strings.Split(result.Logs, "\n") {
		m := line.FindStringSubmatch(raw)
		if m == nil {
			continue
		}
		at, _ := strconv.ParseFloat(m[1], 64)
		msg := m[2]
		switch {
		case strings.HasPrefix(msg, "Casting {SpellID: 15208}") && !strings.Contains(msg, "Cast Time = 0s"):
			castStart["15208"] = at
		case strings.HasPrefix(msg, "Completed cast {SpellID: 15208}"):
			if start, ok := castStart["15208"]; ok {
				casts = append(casts, cast{start, at})
				delete(castStart, "15208")
			}
		case strings.HasPrefix(msg, "Casting {SpellID: 10436}"):
			pulses = append(pulses, at)
		case melee.MatchString(msg):
			swings = append(swings, at)
		}
	}
	if len(casts) < 5 || len(swings) < 5 || len(pulses) < 3 {
		t.Fatalf("%d bolts, %d main-hand swings, %d totem pulses; the sequence did not happen", len(casts), len(swings), len(pulses))
	}

	const eps = 0.011 // the log's two decimals
	held := 0
	for _, at := range swings {
		for _, c := range casts {
			if at > c.start+eps && at < c.end-eps {
				t.Fatalf("a main-hand swing landed at %.2f inside the bolt from %.2f to %.2f", at, c.start, c.end)
			}
			if at > c.end-eps && at < c.end+eps {
				held++ // due during the bolt, it landed as the bolt completed
			}
		}
	}
	if held == 0 {
		t.Fatal("no main-hand swing was held to a bolt's completion; the hold never ran")
	}

	inside := 0
	for _, at := range pulses {
		for _, c := range casts {
			if at > c.start+eps && at < c.end-eps {
				inside++
			}
		}
	}
	if inside == 0 {
		t.Fatal("no totem pulse landed inside a bolt; this build does not exercise the sequence")
	}
}
