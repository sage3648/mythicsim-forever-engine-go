package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/wowsims/forever/tools"
	"github.com/wowsims/forever/tools/database"
)

// A Classic weapon can carry a second damage roll of another school, "+ 1 - 5 Frost Damage" on
// Iceblade Hacker. The Forever client has no such roll: it drops the line, or rebuilds it as an
// equip spell (Iceblade Hacker's "Melee attacks with this weapon deal 41 Frost damage"), and
// Wowhead Forever lists those weapons with their base damage only (Shadowfang 29 - 55, not
// Classic's 29 - 55 plus 4 - 8 Shadow).
//
// The items mergeForeverSimDB fills in come from master's database, which took Wowhead Classic's
// damageMinAll and damageMaxAll: the base roll plus the bonus roll, simmed as Physical. So
// Iceblade Hacker swung for 58 - 111 and its equip spell's Frost damage came on top. This puts
// those weapons back on their base roll, read from the same Wowhead Classic planner.
func dropClassicBonusDamage(db *database.WowDatabase, filled []int32, classicPlannerPath string) {
	rows := readPlannerStats(classicPlannerPath, "Classic gear planner")
	for _, id := range filled {
		item := db.Items[id]
		if item == nil || len(item.ScalingOptions) == 0 || item.ScalingOptions[0] == nil {
			continue
		}
		opt := item.ScalingOptions[0]
		damage, ok := plannerWeaponDamage(fmt.Sprint(id), rows[fmt.Sprint(id)])
		if !ok {
			continue
		}
		if lo, hi, ok := damage.withoutBonus(opt.WeaponDamageMin, opt.WeaponDamageMax); ok {
			opt.WeaponDamageMin, opt.WeaponDamageMax = lo, hi
		}
	}
}

// A weapon's two damage figures in Wowhead's planner: the base roll, and the base plus bonus
// rolls together.
type classicWeaponDamage struct {
	BaseMin, BaseMax float64 // dmgmin1, dmgmax1, unrounded
	AllMin, AllMax   float64 // damageMinAll, damageMaxAll
}

// The base roll, when lo - hi is the base plus a bonus roll. A bonus roll is at least 1 at each
// end; the planner's unrounded base (33.5 for Blade of Eternal Darkness's 34) can sit half a point
// off its rounded total without there being one.
func (d classicWeaponDamage) withoutBonus(lo, hi float64) (float64, float64, bool) {
	if lo != d.AllMin || hi != d.AllMax {
		return 0, 0, false
	}
	if d.AllMin-d.BaseMin < 1 || d.AllMax-d.BaseMax < 1 {
		return 0, 0, false
	}
	return math.Round(d.BaseMin), math.Round(d.BaseMax), true
}

func plannerWeaponDamage(id string, stats map[string]json.RawMessage) (classicWeaponDamage, bool) {
	var d classicWeaponDamage
	for key, field := range map[string]*float64{
		"dmgmin1": &d.BaseMin, "dmgmax1": &d.BaseMax, "damageMinAll": &d.AllMin, "damageMaxAll": &d.AllMax,
	} {
		value, ok := plannerStat(id, stats, key)
		if !ok {
			return d, false
		}
		*field = value
	}
	return d, true
}

// Wowhead's gear planner page data: item id to the item's stats. Most stats are numbers, but some
// keys hold objects (appearances, skillBuff), so they are decoded loosely and read with plannerStat.
func readPlannerStats(path string, what string) map[string]map[string]json.RawMessage {
	text := tools.ReadFile(path)
	start, end := strings.Index(text, "{"), strings.Index(text, "});")
	if start < 0 || end < start {
		panic("invalid " + what)
	}
	raw := strings.TrimSpace(text[start : end+1])
	raw = strings.TrimSuffix(strings.TrimSpace(strings.TrimSuffix(raw, "}")), ",") + "}"
	var rows map[string]struct {
		Stats map[string]json.RawMessage `json:"stats"`
	}
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		panic(err)
	}
	stats := make(map[string]map[string]json.RawMessage, len(rows))
	for id, row := range rows {
		stats[id] = row.Stats
	}
	return stats
}

func plannerStat(id string, stats map[string]json.RawMessage, key string) (float64, bool) {
	value, ok := stats[key]
	if !ok {
		return 0, false
	}
	var number float64
	if err := json.Unmarshal(value, &number); err != nil {
		panic(fmt.Sprintf("gear planner item %s: %q is not a number: %s", id, key, value))
	}
	return number, true
}
