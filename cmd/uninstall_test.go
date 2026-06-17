package cmd

import "testing"

func TestResolveUninstallMode(t *testing.T) {
	tests := []struct {
		name        string
		purge       bool
		interactive bool
		want        uninstallMode
	}{
		{"purge removes everything without prompting", true, true, uninstallMode{prompt: false, removeConfig: true}},
		{"purge wins even non-interactive", true, false, uninstallMode{prompt: false, removeConfig: true}},
		{"interactive asks", false, true, uninstallMode{prompt: true, removeConfig: false}},
		{"non-interactive keeps config silently", false, false, uninstallMode{prompt: false, removeConfig: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveUninstallMode(tt.purge, tt.interactive); got != tt.want {
				t.Fatalf("resolveUninstallMode(%v, %v) = %+v, want %+v", tt.purge, tt.interactive, got, tt.want)
			}
		})
	}
}
