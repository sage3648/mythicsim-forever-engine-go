import { ClassicPhase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	Flask,
	Food,
	IndividualBuffs,
	Profession,
	RaidBuffs,
	SaygesFortune,
	SpellPowerBuff,
	StrengthBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
	SapperExplosive,
} from '../core/proto/common.js';
import { RogueOptions } from '../core/proto/rogue.js';
import { SavedTalents } from '../core/proto/ui.js';
import BackstabAPL from './apls/combat_backstab.apl.json';
import BackstabSweatyAPL from './apls/combat_backstab_sweaty.apl.json';
import SinisterStrikeAPL from './apls/combat_sinister_strike.apl.json';
import SinisterStrikeSweatyAPL from './apls/combat_sinister_strike_sweaty.apl.json';
import SinisterStrikeIEAAPL from './apls/combat_sinister_strike_iea.apl.json';
import HemorrhageAPL from './apls/forever_hemorrhage.apl.json';
import MutilateAPL from './apls/forever_mutilate.apl.json';
import BlankGear from './gear_sets/blank.gear.json';
import BackstabGearLaunch from './gear_sets/backstab_launch.gear.json';
import SinisterStrikeGearLaunch from './gear_sets/sinister_strike_launch.gear.json';
import BackstabGearPreBiS from './gear_sets/combat_backstab_prebis.gear.json';
import SinisterStrikeGearPreBiS from './gear_sets/combat_sinister_strike_prebis.gear.json';
import BackstabGearP1BiS from './gear_sets/combat_backstab_p1_bis.gear.json';
import BackstabGearP2BiS from './gear_sets/combat_backstab_p2_bis.gear.json';
import SinisterStrikeGearP1BiS from './gear_sets/combat_sinister_strike_p1_bis.gear.json';
import SinisterStrikeGearP2BiS from './gear_sets/combat_sinister_strike_p2_bis.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const GearBlank = PresetUtils.makePresetGear('Blank', BlankGear);
export const GearBackstabLaunch = PresetUtils.makePresetGear('Backstab Launch', BackstabGearLaunch);
export const GearSinisterStrikeLaunch = PresetUtils.makePresetGear('Sinister Strike Launch', SinisterStrikeGearLaunch);
export const GearBackstabPreBiS = PresetUtils.makePresetGear('Backstab Pre-BiS', BackstabGearPreBiS);
export const GearSinisterStrikePreBiS = PresetUtils.makePresetGear('Sinister Strike Pre-BiS', SinisterStrikeGearPreBiS);
export const GearBackstabP1BiS = PresetUtils.makePresetGear('Backstab P1 BiS', BackstabGearP1BiS);
export const GearBackstabP2BiS = PresetUtils.makePresetGear('Backstab P2 BiS', BackstabGearP2BiS);
export const GearSinisterStrikeP1BiS = PresetUtils.makePresetGear('Sinister Strike P1 BiS', SinisterStrikeGearP1BiS);
export const GearSinisterStrikeP2BiS = PresetUtils.makePresetGear('Sinister Strike P2 BiS', SinisterStrikeGearP2BiS);

export const GearPresets = {
	[ClassicPhase.Phase1]: [
		GearBackstabLaunch,
		GearSinisterStrikeLaunch,
		GearBackstabPreBiS,
		GearSinisterStrikePreBiS,
		GearBackstabP1BiS,
		GearSinisterStrikeP1BiS,
	],
	[ClassicPhase.Phase2]: [
		GearBackstabLaunch,
		GearSinisterStrikeLaunch,
		GearBackstabPreBiS,
		GearSinisterStrikePreBiS,
		GearBackstabP2BiS,
		GearSinisterStrikeP2BiS,
	],
};

export const DefaultGear = GearSinisterStrikePreBiS;

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets[]
///////////////////////////////////////////////////////////////////////////

export const ROTATION_PRESET_BACKSTAB = PresetUtils.makePresetAPLRotation('Backstab', BackstabAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE = PresetUtils.makePresetAPLRotation('Sinister Strike', SinisterStrikeAPL, {});
export const ROTATION_PRESET_BACKSTAB_SWEATY = PresetUtils.makePresetAPLRotation('Backstab (Sweaty)', BackstabSweatyAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE_SWEATY = PresetUtils.makePresetAPLRotation('Sinister Strike (Sweaty)', SinisterStrikeSweatyAPL, {});
export const ROTATION_PRESET_SINISTER_STRIKE_IEA = PresetUtils.makePresetAPLRotation('Improved Expose Armor (SS)', SinisterStrikeIEAAPL, {});
export const ROTATION_PRESET_MUTILATE = PresetUtils.makePresetAPLRotation('Mutilate', MutilateAPL, {});
export const ROTATION_PRESET_HEMORRHAGE = PresetUtils.makePresetAPLRotation('Hemorrhage', HemorrhageAPL, {});

export const APLPresets = {
	[ClassicPhase.Phase1]: [
		ROTATION_PRESET_BACKSTAB,
		ROTATION_PRESET_SINISTER_STRIKE,
		ROTATION_PRESET_BACKSTAB_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_IEA,
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_HEMORRHAGE,
	],
	[ClassicPhase.Phase2]: [
		ROTATION_PRESET_BACKSTAB,
		ROTATION_PRESET_SINISTER_STRIKE,
		ROTATION_PRESET_BACKSTAB_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_SWEATY,
		ROTATION_PRESET_SINISTER_STRIKE_IEA,
		ROTATION_PRESET_MUTILATE,
		ROTATION_PRESET_HEMORRHAGE,
	],
};

//Need to add main hand equip logic or talent/rotation logic to map to Auto APL
export const DefaultAPLs: Record<number, PresetUtils.PresetRotation> = {
	[0]: ROTATION_PRESET_SINISTER_STRIKE,
	[1]: ROTATION_PRESET_BACKSTAB,
};

export const DefaultAPLBackstab = APLPresets[ClassicPhase.Phase2][0];
export const DefaultAPLSinisterStrike = APLPresets[ClassicPhase.Phase2][1];
export const DefaultAPLIEA = APLPresets[ClassicPhase.Phase2][4];
export const DefaultAPLMutilate = APLPresets[ClassicPhase.Phase2][5];

///////////////////////////////////////////////////////////////////////////
//                                 Talent Presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

// Preset name must be unique. Ex: 'Backstab DPS' cannot be used as a name more than once

export const CombatBackstabTalents = PresetUtils.makePresetTalents('Backstab', SavedTalents.create({ talentsString: '005302005-30230320201515231-102' }));
export const CombatSinisterStrikeTalents = PresetUtils.makePresetTalents(
	'Sinister Strike',
	SavedTalents.create({ talentsString: '00530310501-32003311201515231' }),
);
export const CombatSinisterStrikeIEATalents = PresetUtils.makePresetTalents(
	'Improved Expose Armor (SS)',
	SavedTalents.create({ talentsString: '005303125-32003311201515131' }),
);
export const AssassinationMutilateTalents = PresetUtils.makePresetTalents(
	'Mutilate',
	SavedTalents.create({ talentsString: '00530310521001-302303202015151-002' }),
);

export const TalentsCombatDualWield = PresetUtils.makePresetTalents(
	'Combat Dual-Wield 15/33/3',
	SavedTalents.create({ talentsString: '1053231-22530300001515231-012' }),
);
export const TalentsAssassinationMutilate = PresetUtils.makePresetTalents(
	'Assassination Mutilate 31/20/0',
	SavedTalents.create({ talentsString: '02532010531201051-225303000005' }),
);
export const TalentsSubtletyHemo = PresetUtils.makePresetTalents(
	'Subtlety Hemo 15/0/36',
	SavedTalents.create({ talentsString: '125320101--5320003310013211551' }),
);

export const TalentPresets = {
	[ClassicPhase.Phase1]: [
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		CombatSinisterStrikeIEATalents,
		AssassinationMutilateTalents,
		TalentsCombatDualWield,
		TalentsAssassinationMutilate,
		TalentsSubtletyHemo,
	],
	[ClassicPhase.Phase2]: [
		CombatBackstabTalents,
		CombatSinisterStrikeTalents,
		CombatSinisterStrikeIEATalents,
		AssassinationMutilateTalents,
		TalentsCombatDualWield,
		TalentsAssassinationMutilate,
		TalentsSubtletyHemo,
	],
};

export const DefaultTalentsAssassin = AssassinationMutilateTalents;
export const DefaultTalentsCombat = CombatSinisterStrikeTalents;
export const DefaultTalentsSubtlety = TalentPresets[ClassicPhase.Phase2][0];

export const DefaultTalentsBackstab = TalentPresets[ClassicPhase.Phase2][0];
export const DefaultTalentsSinisterStrike = TalentPresets[ClassicPhase.Phase2][1];
export const DefaultTalentsIEA = TalentPresets[ClassicPhase.Phase2][2];
export const DefaultTalentsMutilate = TalentPresets[ClassicPhase.Phase2][3];

export const DefaultTalents = DefaultTalentsSinisterStrike;

///////////////////////////////////////////////////////////////////////////
//                                Build Presets
///////////////////////////////////////////////////////////////////////////
export const PresetBuildBackstab = PresetUtils.makePresetBuild('Backstab', {
	gear: GearBackstabP2BiS,
	talents: DefaultTalentsBackstab,
	rotation: DefaultAPLBackstab,
});
export const PresetBuildSinisterStrike = PresetUtils.makePresetBuild('Sinister Strike', {
	gear: GearSinisterStrikeP2BiS,
	talents: DefaultTalentsSinisterStrike,
	rotation: DefaultAPLSinisterStrike,
});
export const PresetBuildIEA = PresetUtils.makePresetBuild('IEA', {
	gear: GearSinisterStrikeP2BiS,
	talents: DefaultTalentsIEA,
	rotation: DefaultAPLIEA,
});
export const PresetBuildMutilate = PresetUtils.makePresetBuild('Mutilate', {
	gear: GearBackstabP2BiS,
	talents: DefaultTalentsMutilate,
	rotation: DefaultAPLMutilate,
});

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = RogueOptions.create({});

///////////////////////////////////////////////////////////////////////////
//                         Consumes/Buffs/Debuffs
///////////////////////////////////////////////////////////////////////////

export const P1Consumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredRogueThistleTea,
	dragonBreathChili: true,
	flask: Flask.FlaskOfSupremePower,
	food: Food.FoodGrilledSquid,
	mainHandImbue: WeaponImbue.InstantPoison,
	offHandImbue: WeaponImbue.DeadlyPoison,
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.GroundScorpokAssay,
	sapperExplosive: SapperExplosive.SapperGoblinSapper,
});

export const DefaultConsumes = {
	[ClassicPhase.Phase1]: P1Consumes,
};

export const P1RaidBuffs = RaidBuffs.create({
	battleShout: TristateEffect.TristateEffectImproved,
	fireResistanceAura: true,
	fireResistanceTotem: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	graceOfAirTotem: TristateEffect.TristateEffectImproved,
	leaderOfThePack: true,
	trueshotAura: true,
});

export const DefaultRaidBuffs = {
	[ClassicPhase.Phase1]: P1RaidBuffs,
};

export const P1IndividualBuffs = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	//saygesFortune: SaygesFortune.SaygesDamage,
	spiritOfZandalar: false,
});

export const DefaultIndividualBuffs = {
	[ClassicPhase.Phase1]: P1IndividualBuffs,
};

export const P1DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	faerieFire: true,
	sunderArmor: true,
});

export const DefaultDebuffs = {
	[ClassicPhase.Phase1]: P1DefaultDebuffs,
};

export const P1OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.ProfessionUnknown,
};

export const OtherDefaults = {
	[ClassicPhase.Phase1]: P1OtherDefaults,
};
