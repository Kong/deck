package lua_test

import (
	"testing"

	"github.com/kong/deck/plugin/lua"
)

func TestMultiEditionValidator(t *testing.T) {
	// Test OSS Edition
	t.Run("Kong OSS Validation", func(t *testing.T) {
		v, err := lua.NewValidator("oss", "")
		if err != nil {
			t.Fatalf("Failed to initialize OSS validator: %v", err)
		}
		if v.GetEdition() != "OSS" {
			t.Errorf("Expected edition OSS, got %s", v.GetEdition())
		}

		// OSS should block all 'os.' calls including execute
		code := `os.execute("ls")`
		violations, _ := v.Validate(code, "standard")
		if len(violations) == 0 {
			t.Error("Expected violation for 'os.execute' in OSS, but got none")
		}
	})

	// Test EE Edition
	t.Run("Kong EE Validation", func(t *testing.T) {
		v, err := lua.NewValidator("ee", "")
		if err != nil {
			t.Fatalf("Failed to initialize EE validator: %v", err)
		}

		// EE should detect evasion via _G
		code := `local x = _G["require"]`
		violations, _ := v.Validate(code, "strict")
		found := false
		for _, v := range violations {
			if v.ID == "LUA-EV-001" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected LUA-EV-001 violation in EE, but got none")
		}
	})
}

func TestInsidiousCases(t *testing.T) {
	v, err := lua.NewValidator("ee", "")
	if err != nil {
		t.Fatalf("Failed to initialize validator: %v", err)
	}

	tests := []struct {
		name    string
		code    string
		profile string
	}{
		{
			name: "Whitespace and comments between identifiers",
			code: `kong  .  -- sneaky comment
			log  .
			err("test")`,
			profile: "lua", // Should fail because 'kong' is not in lua profile
		},
		{
			name: "Deeply nested forbidden call",
			code: `if true then
			for i=1,10 do
			if i == 5 then os.execute("ls") end
			end
			end`,
			profile: "lax", // os.execute is forbidden even in lax
		},
		{
			name:    "Forbidden call as function argument",
			code:    `print(os.execute("ls"))`,
			profile: "strict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations, err := v.Validate(tt.code, tt.profile)
			if err != nil {
				t.Fatalf("Engine error: %v", err)
			}

			if len(violations) == 0 {
				t.Errorf("Case '%s' failed: expected violations but got none", tt.name)
			}
		})
	}
}
