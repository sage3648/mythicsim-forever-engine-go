//go:build with_db

package core

import (
	"testing"
)

// Classic's bonus damage roll ("+ 1 - 5 Frost Damage") is not in Forever, so a weapon filled in
// from master's database swings for its base roll only (tools/database/gen_db/classic_bonus_damage.go).
// Checks the embedded binary used by real simulations.
func TestForeverWeaponsLeaveOutClassicBonusDamage(t *testing.T) {
	for id, want := range map[int32][2]float64{
		13952: {57, 106},  // Iceblade Hacker, not 58 - 111
		13982: {142, 214}, // Warblade of Caer Darrow, not 143 - 236
		16039: {129, 194}, // Ta'Kierthan Songblade, not 130 - 214
		1482:  {29, 55},   // Shadowfang, from the client
		19019: {44, 115},  // Thunderfury, from the client
	} {
		item := NewItem(ItemSpec{ID: id})
		if got := [2]float64{item.WeaponDamageMin, item.WeaponDamageMax}; got != want {
			t.Errorf("%d %s weapon damage = %v, want %v", id, item.Name, got, want)
		}
	}
}
