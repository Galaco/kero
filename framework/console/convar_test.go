package console

import "testing"

func TestSetConvarBooleanInt(t *testing.T) {
	// Setup: Create a boolean convar
	AddConvarBool("test_bool", "Test boolean convar", false)

	// Test: Set to 1 (should become true)
	SetConvarBooleanInt("test_bool", 1)
	if !GetConvarBoolean("test_bool") {
		t.Error("Expected true when setting to 1")
	}

	// Test: Set to 0 (should become false)
	SetConvarBooleanInt("test_bool", 0)
	if GetConvarBoolean("test_bool") {
		t.Error("Expected false when setting to 0")
	}

	// Test: Set to other value (should be ignored, remain false)
	SetConvarBooleanInt("test_bool", 2)
	if GetConvarBoolean("test_bool") {
		t.Error("Expected false when setting to 2 (should be ignored)")
	}

	// Test: Set to 1 again
	SetConvarBooleanInt("test_bool", 1)
	if !GetConvarBoolean("test_bool") {
		t.Error("Expected true when setting to 1")
	}

	// Test: Set to -1 (should be ignored)
	SetConvarBooleanInt("test_bool", -1)
	if !GetConvarBoolean("test_bool") {
		t.Error("Expected true (unchanged) when setting to -1 (should be ignored)")
	}
}
