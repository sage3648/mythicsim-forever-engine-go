package core

import (
	"math"
	"testing"
)

func TestFeralWeaponFloat32Formula(t *testing.T) {
	for _, tc := range []struct {
		name                               string
		dps                                float32
		ms                                 int32
		variance                           float32
		ap, low, high, sheetLow, sheetHigh float64
	}{
		{"Heartseeker cat", 41.49628067017, 1000, .4, 1000, 33, 50, 104, 122},
		{"Impervious Giant bear", 49.02209854126, 2500, .4, 1000, 98, 147, 276, 326},
		{"unarmed cat", .5, 1000, 0, 0, 1, 1, 1, 1},
		{"unarmed bear", .5, 2500, 0, 0, 1, 1, 1, 1},
		{"sub-one DPS clamp", .1, 1000, .4, 14, 1, 1, 2, 2},
		{"fractional AP", 54.80965423584, 2500, .4, 1234.567, 109, 164, 329, 385},
		{"cat float32 rounding boundary", 2.0833332538604736, 1000, .4, 0, 1, 3, 1, 3},
		{"bear float32 rounding boundary", .8333333134651184, 2500, .4, 0, 1, 3, 1, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			low, high := FeralWeaponBaseDamage(tc.dps, tc.ms, tc.variance)
			if low != tc.low || high != tc.high {
				t.Fatalf("base = %v..%v, want %v..%v", low, high, tc.low, tc.high)
			}
			low, high = FeralWeaponCharacterSheetDamage(tc.dps, tc.ms, tc.variance, tc.ap)
			if low != tc.sheetLow || high != tc.sheetHigh {
				t.Fatalf("sheet = %v..%v, want %v..%v", low, high, tc.sheetLow, tc.sheetHigh)
			}
		})
	}
}

func TestFeralWeaponKeepsCombatAPSeparateFromDisplayRounding(t *testing.T) {
	w := NewFeralWeapon(41.49628067017, 1000, .4)
	if got := w.CalculateAverageWeaponDamage(1000); got != 112.92857360839844 {
		t.Fatalf("combat mean = %.15f, want unquantized float32 AP contribution", got)
	}
	if w.SwingSpeed != 1 || w.NormalizedSwingSpeed != 1 {
		t.Fatal("cat weapon uses equipped weapon speed")
	}
	bear := NewFeralWeapon(49.02209854126, 2500, .4)
	if bear.SwingSpeed != 2.5 || bear.NormalizedSwingSpeed != 2.5 {
		t.Fatal("bear speed changed")
	}
	if got := FeralWeaponAttackPowerDamage(1000, 2500); got != 178.57144165039062 {
		t.Fatalf("AP millisecond order: %v", got)
	}
	// Other classes retain their existing float64 weapon calculations.
	ordinary := Weapon{BaseDamageMin: 10, BaseDamageMax: 20, SwingSpeed: 2, AttackPowerPerDPS: 14}
	if math.Abs(ordinary.CalculateAverageWeaponDamage(1000)-(15+2000.0/14)) > 1e-12 {
		t.Fatal("ordinary weapon changed")
	}
}
