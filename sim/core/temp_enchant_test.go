package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// A stone or oil in the temporary enchantment slot counts like one in the imbue slot, beside whatever
// the imbue slot holds (mythicsim patch 100).
func TestTempEnchantStacksBesideTheImbue(t *testing.T) {
	crit := consumesStatDelta(t, stats.PhysicalCritPercent, func(request *proto.RaidSimRequest) {
		consumesOf(request).MhImbueId = 25121 // Wizard Oil in the imbue slot
		consumesOf(request).MhTempEnchantId = 22756
	})
	if !WithinToleranceFloat64(2, crit, 0.001) {
		t.Fatalf("an Elemental Sharpening Stone beside an imbue should grant 2%% melee crit, got %0.3f%%", crit)
	}
	spell := consumesStatDelta(t, stats.SpellDamage, func(request *proto.RaidSimRequest) {
		consumesOf(request).MhImbueId = 22756
		consumesOf(request).MhTempEnchantId = 25121
	})
	if !WithinToleranceFloat64(24, spell, 0.01) {
		t.Fatalf("Wizard Oil beside an imbue should grant 24 spell damage, got %0.2f", spell)
	}

	_, offHand := setupConsumesSim(func(request *proto.RaidSimRequest) {
		consumesOf(request).OhTempEnchantId = 16138
	})
	if offHand.AutoAttacks.OH().BaseDamageMax != 8 || offHand.AutoAttacks.MH().BaseDamageMax != 0 {
		t.Fatalf("an off-hand Dense Sharpening Stone should add 8 to the off hand only, got %0.2f / %0.2f",
			offHand.AutoAttacks.OH().BaseDamageMax, offHand.AutoAttacks.MH().BaseDamageMax)
	}
	_, mainHand := setupConsumesSim(func(request *proto.RaidSimRequest) {
		consumesOf(request).MhTempEnchantId = 16622
	})
	if mainHand.MHImbueFlatWeaponDamage() != 8 {
		t.Fatalf("the main hand's flat imbue damage should count the temporary enchantment, got %0.2f", mainHand.MHImbueFlatWeaponDamage())
	}
}
