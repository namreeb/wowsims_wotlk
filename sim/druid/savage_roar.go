package druid

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func (druid *Druid) getSavageRoarMultiplier() float64 {
	glyphBonus := core.TernaryFloat64(druid.HasMajorGlyph(proto.DruidMajorGlyph_GlyphOfSavageRoar), 0.03, 0)
	return 1.3 + glyphBonus
}

func (druid *Druid) registerSavageRoarSpell() {
	actionID := core.ActionID{SpellID: 52610}
	baseCost := 25.0

	srm := druid.getSavageRoarMultiplier()
	durTable := druid.SavageRoarDurationTable()

	druid.SavageRoarAura = druid.RegisterAura(core.Aura{
		Label:    "Savage Roar Aura",
		ActionID: actionID,
		Duration: 9,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= srm
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if druid.InForm(Cat) {
				druid.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] /= srm
			}
		},
	})

	srSpell := druid.RegisterSpell(core.SpellConfig{
		ActionID: actionID,

		ResourceType: stats.Energy,
		BaseCost:     baseCost,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				Cost: baseCost,
				GCD:  time.Second,
			},
			IgnoreHaste: true,
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			druid.SavageRoarAura.Duration = durTable[druid.ComboPoints()]
			druid.SavageRoarAura.Activate(sim)
			druid.SpendComboPoints(sim, spell.ComboPointMetrics())
		},
	})

	druid.SavageRoar = srSpell
}

func (druid *Druid) SavageRoarDurationTable() [6]time.Duration {
	durBonus := core.TernaryDuration(druid.setBonuses.feral_t8_4, time.Second*8, 0)
	return [6]time.Duration{
		0,
		durBonus + time.Second*(9+5),
		durBonus + time.Second*(9+10),
		durBonus + time.Second*(9+15),
		durBonus + time.Second*(9+20),
		durBonus + time.Second*(9+25),
	}
}

func (druid *Druid) CanSavageRoar() bool {
	return druid.InForm(Cat) && druid.ComboPoints() > 0 && (druid.CurrentEnergy() >= druid.CurrentSavageRoarCost())
}

func (druid *Druid) CurrentSavageRoarCost() float64 {
	return druid.SavageRoar.ApplyCostModifiers(druid.SavageRoar.BaseCost)
}
