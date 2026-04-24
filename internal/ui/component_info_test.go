package ui

import "testing"

// TestToolInfoCatalogCoversAllPlaceableTools ensures every Tool that can
// appear in the inventory has user-facing copy. A missing entry would
// render an empty description panel in-game; this test makes adding a new
// Tool without description text a compile-green but test-red change.
func TestToolInfoCatalogCoversAllPlaceableTools(t *testing.T) {
	for _, tool := range PlaceableTools {
		info, ok := ToolInfoCatalog[tool]
		if !ok {
			t.Errorf("PlaceableTool %v has no ToolInfoCatalog entry", tool)
			continue
		}
		if info.Name == "" {
			t.Errorf("tool %v: Name is empty", tool)
		}
		if info.Description == "" {
			t.Errorf("tool %v: Description is empty", tool)
		}
	}
}

func TestToolInfoCatalogIncludesErase(t *testing.T) {
	// Erase isn't in PlaceableTools (it's a right-click affordance), but
	// the header label is still consumed by the compact palette, so the
	// catalog should include it.
	if _, ok := ToolInfoCatalog[ToolErase]; !ok {
		t.Fatal("ToolErase should have a catalog entry")
	}
}

func TestKindForToolMatchesPlaceable(t *testing.T) {
	for _, tool := range PlaceableTools {
		if KindForTool(tool) == "" {
			t.Errorf("PlaceableTool %v: KindForTool returned empty", tool)
		}
	}
}
