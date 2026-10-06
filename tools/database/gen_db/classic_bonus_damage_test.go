package main

import (
	"fmt"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database"
)

const classicPlanner = "../../../assets/db_inputs/wowhead_gearplannerdb.txt"

func weapon(id int32, lo, hi float64) *proto.UIItem {
	return &proto.UIItem{Id: id, ScalingOptions: map[int32]*proto.ScalingItemProperties{
		0: {WeaponDamageMin: lo, WeaponDamageMax: hi},
	}}
}

func TestDropClassicBonusDamage(t *testing.T) {
	db := database.NewWowDatabase()
	db.Items[13952] = weapon(13952, 58, 111)  // Iceblade Hacker: 57 - 106 + 1 - 5 Frost
	db.Items[13982] = weapon(13982, 143, 236) // Warblade of Caer Darrow: 142 - 214 + 1 - 22 Frost
	db.Items[16039] = weapon(16039, 130, 214) // Ta'Kierthan Songblade: 129 - 194 + 1 - 20 Frost
	db.Items[17780] = weapon(17780, 34, 70)   // Blade of Eternal Darkness: no bonus, planner base 33.5 - 69.5
	db.Items[1482] = weapon(1482, 29, 55)     // Shadowfang from the client: already its base roll
	db.Items[19019] = weapon(19019, 44, 115)  // Thunderfury from the client: already its base roll
	filled := []int32{13952, 13982, 16039, 17780, 1482, 19019}

	dropClassicBonusDamage(db, filled, classicPlanner)

	want := map[int32][2]float64{
		13952: {57, 106},
		13982: {142, 214},
		16039: {129, 194},
		17780: {34, 70},
		1482:  {29, 55},
		19019: {44, 115},
	}
	for id, w := range want {
		opt := db.Items[id].ScalingOptions[0]
		if got := [2]float64{opt.WeaponDamageMin, opt.WeaponDamageMax}; got != w {
			t.Errorf("item %d: weapon damage %v, want %v", id, got, w)
		}
	}
}

// Only the items mergeForeverSimDB filled in are ours to correct: a client row is the client's.
func TestDropClassicBonusDamageLeavesClientRows(t *testing.T) {
	db := database.NewWowDatabase()
	db.Items[13952] = weapon(13952, 58, 111)

	dropClassicBonusDamage(db, nil, classicPlanner)

	opt := db.Items[13952].ScalingOptions[0]
	if got := fmt.Sprint(opt.WeaponDamageMin, " - ", opt.WeaponDamageMax); got != "58 - 111" {
		t.Errorf("client row changed to %s", got)
	}
}
