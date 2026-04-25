package sim

import "particleaccelerator/internal/bignum"

var BinderStoreBaseCapacity = map[Element]int{
	ElementHydrogen: 15,
	ElementHelium:   8,
	ElementLithium:  30,
	ElementCarbon:   100,
}

var BinderStoreElementOrder = []Element{
	ElementHydrogen,
	ElementHelium,
	ElementLithium,
	ElementCarbon,
}

func (s *GameState) EffectiveBinderStoreCapacity(e Element) int {
	base := BinderStoreBaseCapacity[e]
	if base <= 0 {
		return 0
	}
	mul := s.Modifiers.Normalized().BinderStoreCapacityMul
	cap := bignum.FromInt(base).Mul(mul).Float64()
	if cap <= 0 {
		return 0
	}
	if cap < 1 {
		return 1
	}
	return int(cap)
}

func (s *GameState) BankSubject(e Element) bool {
	if _, ok := ElementCatalog[e]; !ok {
		return false
	}
	cap := s.EffectiveBinderStoreCapacity(e)
	if cap <= 0 {
		return false
	}
	if s.BinderReserves == nil {
		s.BinderReserves = map[Element]int{}
	}
	if s.BinderReserves[e] >= cap {
		return false
	}
	s.BinderReserves[e]++
	return true
}

func HasAnyToken(s *GameState) bool {
	if s == nil {
		return false
	}
	for _, n := range s.TokenInventory {
		if n > 0 {
			return true
		}
	}
	return false
}

func HasAnyBond(s *GameState) bool {
	if s == nil {
		return false
	}
	for _, owned := range s.BondsState {
		if owned {
			return true
		}
	}
	return false
}

func CanCrystalliseToken(s *GameState, e Element) bool {
	if s == nil {
		return false
	}
	cost := CrystallisationCost(e, s.TokenInventory[e])
	return cost > 0 && s.BinderReserves[e] >= cost
}

func CanSynthesiseBond(s *GameState, id BondID) bool {
	if s == nil || s.BondsState[id] {
		return false
	}
	bond, ok := BondCatalog[id]
	if !ok {
		return false
	}
	for e, n := range bond.TokenCost {
		if s.TokenInventory[e] < n {
			return false
		}
	}
	return true
}
