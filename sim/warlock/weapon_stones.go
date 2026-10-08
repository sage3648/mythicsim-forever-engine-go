package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Forever's Firestone and Spellstone are weapon imbues rather than held off-hand items, and they stack
// with a temporary enchantment such as Wizard Oil (Blizzard's Rogue and Warlock deep dive, 8 October
// 2026). The imbue is the Create spell's triggered aura at its highest rank, read from client data:
// Firestone (23483) adds Fire damage and spell critical strike chance, Spellstone (1237165) adds damage
// to the schools in its mask (Fire and Shadow at the pinned build, mask 36) and cast speed. Mythicsim
// patch 101.
func (warlock *Warlock) registerWeaponStone() {
	var imbue *spelldata.Spell
	switch warlock.Options.WeaponStone {
	case proto.WarlockOptions_Firestone:
		imbue = spellData.CreateFirestoneTriggered.Highest()
	case proto.WarlockOptions_Spellstone:
		imbue = spellData.CreateSpellstoneTriggered.Highest()
	default:
		return
	}
	bonus := weaponStoneStats(imbue)
	aura := core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label:      imbue.Name,
		ActionID:   core.ActionID{SpellID: imbue.ID},
		BuildPhase: core.CharacterBuildPhaseGear,
		Duration:   core.NeverExpires,
	}))
	aura.AttachStatsBuff(bonus).AttachHastePseudoStats(imbue.SpeedPseudoStats())
}

// The school damage stat each bit of an A_MOD_DAMAGE_DONE school mask raises.
var schoolDamageStats = []struct {
	mask int32
	stat stats.Stat
}{
	{2, stats.HolyDamage}, {4, stats.FireDamage}, {8, stats.NatureDamage},
	{16, stats.FrostDamage}, {32, stats.ShadowDamage}, {64, stats.ArcaneDamage},
}

// weaponStoneStats is the imbue's school damage and spell crit as stats; its cast speed goes through
// the spell's speed pseudo stats.
func weaponStoneStats(imbue *spelldata.Spell) stats.Stats {
	bonus := stats.Stats{}
	for i := range imbue.Effects {
		e := &imbue.Effects[i]
		if e.Type != dbcenums.E_APPLY_AURA {
			continue
		}
		switch e.Aura {
		case dbcenums.A_MOD_DAMAGE_DONE:
			for _, school := range schoolDamageStats {
				if e.Misc&school.mask != 0 {
					bonus[school.stat] += e.BaseValue()
				}
			}
		case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
			bonus[stats.SpellCritPercent] += e.BaseValue()
		}
	}
	return bonus
}
