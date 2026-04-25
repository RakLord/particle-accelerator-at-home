package sim

import (
	"errors"
	"testing"

	"particleaccelerator/internal/bignum"
)

func TestPurchaseLabUpgradeDeductsBPAndRebuildsModifiers(t *testing.T) {
	s := NewGameState()
	s.BondPoints = 3
	if err := PurchaseLabUpgrade(s, LabDensePacking); err != nil {
		t.Fatalf("PurchaseLabUpgrade: %v", err)
	}
	if got := s.BondPoints; got != 2 {
		t.Fatalf("BondPoints after purchase: got %d want 2", got)
	}
	if got := s.LaboratoryUpgrades[LabDensePacking]; got != 1 {
		t.Fatalf("Dense Packing level: got %d want 1", got)
	}
	if got := s.Modifiers.Normalized().BinderStoreCapacityMul; !got.Eq(bignum.FromInt(2)) {
		t.Fatalf("BinderStoreCapacityMul: got %v want 2", got)
	}
	if cap := s.EffectiveBinderStoreCapacity(ElementCarbon); cap != 200 {
		t.Fatalf("Carbon store capacity: got %d want 200", cap)
	}
}

func TestPurchaseLabUpgradeRejectsUnaffordable(t *testing.T) {
	s := NewGameState()
	err := PurchaseLabUpgrade(s, LabStableIsotope)
	if !errors.Is(err, ErrInsufficientBondPoints) {
		t.Fatalf("PurchaseLabUpgrade err = %v want ErrInsufficientBondPoints", err)
	}
	if got := s.LaboratoryUpgrades[LabStableIsotope]; got != 0 {
		t.Fatalf("unaffordable upgrade changed level: got %d", got)
	}
}

func TestResetGenesisPreservesPrestigeAndWipesRunState(t *testing.T) {
	s := NewGameState()
	s.USD = bignum.FromInt(999)
	s.Research[ElementHydrogen] = 20
	s.Research[ElementCarbon] = 10
	s.UnlockedElements[ElementHelium] = true
	s.UnlockedElements[ElementLithium] = true
	s.Grid.Cells[0][0].IsCollector = true
	s.Owned[KindCollector] = 4
	s.CurrentLoad = 1
	s.InjectionCooldownRemaining = 9
	s.BinderReserves[ElementCarbon] = 7
	s.TokenInventory[ElementCarbon] = 1
	s.BondsState[BondMethane] = true
	s.BondPoints = 5
	s.LaboratoryUpgrades[LabStableIsotope] = 1
	s.LaboratoryUpgrades[LabCarbonCore] = 1
	s.AutoInjectActive = true
	s.BestStats[ElementHydrogen] = ElementBestStats{MaxSpeed: SpeedFromInt(4)}

	ResetGenesis(s)

	if !s.USD.IsZero() {
		t.Fatalf("USD after reset: got %v want 0", s.USD)
	}
	if s.Grid.Cells[0][0].IsCollector {
		t.Fatalf("grid layout was not wiped")
	}
	if got := s.Owned[KindCollector]; got != 1 {
		t.Fatalf("Owned collectors: got %d want starter 1", got)
	}
	if got := s.Research[ElementHydrogen]; got != 6 {
		t.Fatalf("Hydrogen research carryover: got %d want 6", got)
	}
	if got := s.Research[ElementCarbon]; got != 3 {
		t.Fatalf("Carbon research carryover: got %d want 3", got)
	}
	for _, e := range []Element{ElementHydrogen, ElementHelium, ElementLithium, ElementBeryllium, ElementBoron, ElementCarbon} {
		if !s.UnlockedElements[e] {
			t.Fatalf("%s should be unlocked by Carbon Core", e)
		}
	}
	if s.UnlockedElements[ElementNitrogen] {
		t.Fatalf("Carbon Core should not unlock Nitrogen")
	}
	if len(s.BinderReserves) != 0 || len(s.TokenInventory) != 0 {
		t.Fatalf("Binder/token run state not wiped: reserves=%v tokens=%v", s.BinderReserves, s.TokenInventory)
	}
	if !s.BondsState[BondMethane] || s.BondPoints != 5 || !s.AutoInjectActive {
		t.Fatalf("prestige fields not preserved: bonds=%v BP=%d auto=%v", s.BondsState, s.BondPoints, s.AutoInjectActive)
	}
	if _, ok := s.BestStats[ElementHydrogen]; !ok {
		t.Fatalf("BestStats should persist")
	}
	if s.CurrentLoad != 0 || s.InjectionCooldownRemaining != 0 || s.RunCount != 1 {
		t.Fatalf("reset counters wrong: load=%d cooldown=%d run=%d", s.CurrentLoad, s.InjectionCooldownRemaining, s.RunCount)
	}
}
