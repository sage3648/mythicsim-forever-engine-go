package core

import "math"

// FeralWeaponBaseDamage evaluates the community's Forever form-weapon formula.
// Each conversion deliberately forces float32 rounding before the next operation.
// Swing time is in milliseconds, even though Weapon stores seconds.
func FeralWeaponBaseDamage(tableDPS float32, swingTimeMS int32, variance float32) (float64, float64) {
	swing := float32(swingTimeMS)
	average := float32(float32(tableDPS*swing) * float32(0.001))
	halfVariance := float32(variance / float32(2))
	low := float32(average * float32(float32(1)-halfVariance))
	high := float32(average * float32(float32(1)+halfVariance))
	return math.Max(1, math.Floor(float64(low))), math.Max(1, math.Round(float64(high)))
}

// FeralWeaponAttackPowerDamage retains the specified millisecond operation order.
func FeralWeaponAttackPowerDamage(attackPower float64, swingTimeMS int32) float64 {
	perMS := float32(float32(0.001) / float32(14))
	perAP := float32(float32(attackPower) * perMS)
	return float64(float32(perAP * float32(swingTimeMS)))
}

// FeralWeaponCharacterSheetDamage applies the final display rounding separately
// from the pre-AP weapon damage. The supplied formula describes these last two
// roundings as character-sheet behavior, not combat-roll quantization.
func FeralWeaponCharacterSheetDamage(tableDPS float32, swingTimeMS int32, variance float32, attackPower float64) (float64, float64) {
	low, high := FeralWeaponBaseDamage(tableDPS, swingTimeMS, variance)
	ap := float32(FeralWeaponAttackPowerDamage(attackPower, swingTimeMS))
	return math.Floor(float64(float32(float32(low) + ap))), math.Ceil(float64(float32(float32(high) + ap)))
}

// NewFeralWeapon creates a form weapon from unrounded client table DPS. Callers
// must resolve caster-weapon adjustments before passing tableDPS.
func NewFeralWeapon(tableDPS float32, swingTimeMS int32, variance float32) Weapon {
	low, high := FeralWeaponBaseDamage(tableDPS, swingTimeMS, variance)
	speed := float64(swingTimeMS) / 1000
	return Weapon{
		BaseDamageMin: low, BaseDamageMax: high,
		SwingSpeed: speed, NormalizedSwingSpeed: speed,
		AttackPowerPerDPS: DefaultAttackPowerPerDPS, MaxRange: MaxMeleeRange,
		feralSwingTimeMS: swingTimeMS,
	}
}
