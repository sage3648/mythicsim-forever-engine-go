package druid

import (
	"github.com/wowsims/forever/sim/core"
)

var rakeRank = spellData.Rake.Highest()
var rakeTick = rakeRank.PeriodicEffect()

// Rake adds 5.26% of attack power to the hit and to every tick. The client rows at the pinned build
// carry no BonusCoefficientFromAP on either, and Blizzard's Druid deep dive (30 September 2026,
// worldofwarcraft.blizzard.com/en-us/news/24301515) only says "Rake: Now gains increased damage from
// Attack Power", so the share is a fit to beta combat logs (Hameru, MythicSim Discord #bugs, 7 October
// 2026, level 30, non-crits on many mobs): at 324 attack power Rake hit 40 and ticked 33, 33, 33; at
// 225 it hit 35 and ticked 28, 28, 27. That rank's base is 23 on the hit and 16 a tick, which leaves
// 0.0526 of attack power on both. The share rides on whatever rank is cast, which keeps its own base
// from client data. A tick reads the share from the attack power it has then, as Rip's does.
const rakeAttackPowerShare = 0.0526

func rakeHitDamage(attackPower float64) float64 {
	return rakeRank.DamageEffect().Average(core.CharacterLevel) + rakeAttackPowerShare*attackPower
}

func rakeTickDamage(attackPower float64) float64 {
	return rakeTick.Average(core.CharacterLevel) + rakeAttackPowerShare*attackPower
}

func (druid *Druid) registerRakeSpell() {
	druid.Rake = druid.RegisterSpell(Cat, core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rakeRank.ID},
		SpellSchool:    rakeRank.SpellSchool(),
		DefenseType:    rakeRank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		ClassSpellMask: DruidSpellRake,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		Rank:           rakeRank.RankNumber(),

		EnergyCost: core.EnergyCostOptions{
			Cost:   int32(rakeRank.Cost()),
			Refund: rakeRank.MissRefund(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rakeRank.GCD(),
			},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		MaxRange:         core.MaxMeleeRange,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    "Rake",
				Duration: rakeRank.Duration(),
			},
			NumberOfTicks: int32(rakeRank.Duration() / rakeTick.Period()),
			TickLength:    rakeTick.Period(),

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.SnapshotPhysical(target, rakeTickDamage(dot.Spell.MeleeAttackPower(target)))
				dot.SnapshotAttackPowerShare(target, rakeAttackPowerShare, false)
				druid.UpdateBleedPower(druid.Rake, sim, target, true, true)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, periodicTickOutcome(rakeRank, dot))
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, rakeHitDamage(spell.MeleeAttackPower(target)), spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if result.Landed() {
				druid.AddComboPoints(sim, 1, spell.ComboPointMetrics())
				spell.Dot(target).Apply(sim)
			} else {
				spell.IssueRefund(sim)
			}
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			if useSnapshot {
				dot := spell.Dot(target)
				return dot.CalcSnapshotDamage(sim, target, dot.OutcomeTick)
			}
			ticks := spell.CalcPeriodicDamage(sim, target, rakeTickDamage(spell.MeleeAttackPower(target)), spell.OutcomeExpectedMagicAlwaysHit)
			attackTable := spell.Unit.AttackTables[target.UnitIndex]
			critChance := spell.PhysicalCritChance(attackTable)
			ticks.Damage *= 1 + critChance*(spell.CritDamageMultiplier(attackTable)-1)
			return ticks
		},
	})

	druid.Rake.ShortName = "Rake"
}

func (druid *Druid) CurrentRakeCost() float64 {
	return druid.Rake.Cost.GetCurrentCost()
}
