package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// StriderKickSpeedsUp (MythicSim patch 120) gives Strider Kick its movement speed: the client row's third
// effect (1317257 effect 2, A_MOD_INCREASE_SPEED 30 on the caster for the spell's 3 sec), and the Sep 24
// beta notes ("Strider Kick now increases movement speed by 30% for 3 seconds."). Without it the Hunter
// walks out of melee at the base 7 yards a second after a kick; it only matters to a rotation that moves.
// Set it false for the kick without the speed.
var StriderKickSpeedsUp = true

// StrikesKeepSlowWeaponSpeed (MythicSim patch 187): Strider Kick and Mongoose Bite take the main hand's real speed
// when it is slower than the normalized one, so a slow weapon keeps its speed and a fast one is lifted to the
// normalized value (the higher of the two). Two sets of Forever logs: a 2.8 and a 1.7 one-hander on the same mobs
// (warcraftlogs wT6F3K4LMWBgzbfP, 30 non-crit Strider Kicks: the 2.8 fits its real speed, the 1.7 the normalized
// 2.4), and a 3.8 and a 2.9 two-hander (Par9kghHbAVYptwm: Strider Kick lands 19.7 above the 3.8's white hits and
// 19.3 above the 2.9's; at 3.3 for both the 3.8 would sit about 23 lower). False restores the normalized speed.
var StrikesKeepSlowWeaponSpeed = true

// strikeWeaponDamage is the weapon part of Strider Kick and Mongoose Bite.
func (hunter *Hunter) strikeWeaponDamage(sim *core.Simulation, attackPower float64) float64 {
	mh := hunter.AutoAttacks.MH()
	if StrikesKeepSlowWeaponSpeed && mh.SwingSpeed > mh.NormalizedSwingSpeed {
		return mh.CalculateWeaponDamage(sim, attackPower)
	}
	return mh.CalculateNormalizedWeaponDamage(sim, attackPower)
}

// Strider Kick from the beta client (1317257): 100% normalized melee weapon damage, 8 sec cooldown,
// 5.81% of base mana.
func (hunter *Hunter) registerStriderKickSpell() {
	if !hunter.Talents.StriderKick {
		return
	}

	rank := spellData.StriderKick.Highest()

	var speed *core.Aura
	if e := rank.Effect(dbcenums.A_MOD_INCREASE_SPEED, 0); StriderKickSpeedsUp && e != nil {
		speed = hunter.RegisterAura(core.Aura{
			Label:    "Strider Kick",
			ActionID: core.ActionID{SpellID: rank.ID},
			Duration: rank.Duration(),
		})
		speed.NewActiveMovementSpeedEffect(e.Percent())
	}

	hunter.StriderKick = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ClassSpellMask: HunterSpellStriderKick,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		MaxRange:       float64(rank.MaxRange),

		ManaCost: rank.ManaCost(),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rank.GCD(),
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			damage := hunter.strikeWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			if speed != nil {
				speed.Activate(sim)
			}
		},
	})
}
