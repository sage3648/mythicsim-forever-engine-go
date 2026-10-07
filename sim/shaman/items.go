package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var totemOfThunder = spelldata.MustFind(461295)

func init() {

	// 	aura := core.MakePermanent(character.RegisterAura(core.Aura{
	// 		Label: "Increased Shock Damage",
	// 	}).AttachSpellMod(core.SpellModConfig{
	// 		Kind:       core.SpellMod_BaseDamage_Flat,
	// 		FloatValue: 30.0,
	// 		ClassMask:  SpellMaskShock,
	// 	}))

	// 	character.ItemSwap.RegisterProc(22395, aura)
	// })

	// Totem of the Storm
	core.NewItemEffect(23199, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increased Lightning Damage",
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: 33.0,
			ClassMask:  SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskOverload,
		}))

		character.ItemSwap.RegisterProc(23199, aura)
	})

	// Totem of Thunder
	// https://www.wowhead.com/forever/item=228176/totem-of-thunder
	//
	// Equip: Increases the critical strike chance of Lightning Bolt by 1% (461295: A_ADD_FLAT_MODIFIER misc 7,
	// crit chance, on class mask word 0 bit 0). Every Lightning Bolt rank carries that bit, and so does each
	// Lightning Overload row (408439 to 408477), so the overload's second bolt takes the 1% too. Chain
	// Lightning and its overload do not carry it.
	core.NewItemEffect(228176, func(agent core.Agent) {
		character := agent.GetCharacter()

		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label:    "Totem of Thunder",
			ActionID: core.ActionID{SpellID: totemOfThunder.ID},
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Percent,
			FloatValue: totemOfThunder.EffectN(1).BaseValue(),
			ClassMask:  SpellMaskLightningBolt | SpellMaskLightningBoltOverload,
		}))

		character.ItemSwap.RegisterProc(228176, aura)
	})

	// Wushoolay's Charm of Spirits
	// https://www.wowhead.com/forever/item=19956/wushoolays-charm-of-spirits
	//
	// Use: Increases the damage dealt by your Lightning Shield spell by 100% for 20 sec (24499).
	// 3 min cooldown, 20 sec on the burst trinket category. The client's mod is additive (aura 108),
	// so it adds to Improved Lightning Shield rather than doubling the total the way Classic's did.
	core.NewItemEffect(19956, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		duration := time.Second * 20

		aura := shaman.RegisterAura(core.Aura{
			Label:    "Energized Shield",
			ActionID: core.ActionID{SpellID: 24499},
			Duration: duration,
		}).AttachSpellMod(core.SpellModConfig{
			Kind:       core.SpellMod_DamageDone_Flat,
			FloatValue: 1,
			ClassMask:  SpellMaskLightningShield,
		})

		spell := shaman.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 19956},
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    shaman.NewTimer(),
					Duration: time.Minute * 3,
				},
				SharedCD: core.Cooldown{
					Timer:    shaman.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				aura.Activate(sim)
			},
		})

		shaman.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
		})
	})
}
