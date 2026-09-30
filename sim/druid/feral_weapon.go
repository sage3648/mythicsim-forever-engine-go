package druid

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/wowsims/forever/sim/core"
)

// Candidate data intentionally rejects unresolved caster weapons instead of
// silently interpreting rounded tooltip DPS as exact client table DPS.
//
//go:embed data/feral_weapon_candidates.json
var feralWeaponData []byte

var feralWeaponCandidates = func() map[int32]struct {
	TableDPS   float32 `json:"tableDps"`
	Unresolved string  `json:"unresolved"`
} {
	var data struct {
		Weapons map[int32]struct {
			TableDPS   float32 `json:"tableDps"`
			Unresolved string  `json:"unresolved"`
		} `json:"weapons"`
	}
	if err := json.Unmarshal(feralWeaponData, &data); err != nil {
		panic(err)
	}
	return data.Weapons
}()

func (druid *Druid) formWeapon(swingTimeMS int32) core.Weapon {
	item := druid.GetMHWeapon()
	if item == nil {
		return core.NewFeralWeapon(.5, swingTimeMS, 0)
	}
	data, ok := feralWeaponCandidates[item.ID]
	if !ok || data.Unresolved != "" {
		panic(fmt.Sprintf("Feral weapon scaling candidate: item %d has no verified effective table DPS (%s)", item.ID, data.Unresolved))
	}
	return core.NewFeralWeapon(data.TableDPS, swingTimeMS, .4)
}
