package cmd

import (
	"path/filepath"
	"testing"
)

func TestBuildInstallPlanInteractiveNoFlags(t *testing.T) {
	plan, err := buildInstallPlan(false, -1, "", "", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Interactive || plan.InstallChromium || plan.InstallSkill {
		t.Fatalf("plan = %+v, want interactive only", plan)
	}
}

func TestBuildInstallPlanNonInteractiveNoFlagsInstallsChromium(t *testing.T) {
	plan, err := buildInstallPlan(false, -1, "", "", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Interactive || !plan.InstallChromium || plan.InstallSkill {
		t.Fatalf("plan = %+v, want chromium only", plan)
	}
}

func TestBuildInstallPlanChromiumFlag(t *testing.T) {
	plan, err := buildInstallPlan(true, -1, "", "", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Interactive || !plan.InstallChromium || plan.Revision != -1 {
		t.Fatalf("plan = %+v, want direct chromium install", plan)
	}
}

func TestBuildInstallPlanChromiumLegacyRevisionArg(t *testing.T) {
	plan, err := buildInstallPlan(true, -1, "", "", []string{"123"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.InstallChromium || plan.Revision != 123 {
		t.Fatalf("plan = %+v, want chromium revision 123", plan)
	}
}

func TestBuildInstallPlanSkillTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	plan, err := buildInstallPlan(false, -1, "claude", "", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Interactive || plan.InstallChromium || !plan.InstallSkill || plan.SkillTarget != "claude" {
		t.Fatalf("plan = %+v, want direct claude skill install", plan)
	}

	plan, err = buildInstallPlan(false, -1, "all", "", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.InstallSkill || plan.SkillTarget != "all" {
		t.Fatalf("plan = %+v, want all skills", plan)
	}
}

func TestBuildInstallPlanSkillPath(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "skills")
	plan, err := buildInstallPlan(false, -1, "codex", parent, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.InstallSkill || plan.SkillTarget != "codex" || plan.SkillPath != parent {
		t.Fatalf("plan = %+v, want codex skill with custom path", plan)
	}
}

func TestBuildInstallPlanRejectsInvalidCombos(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if _, err := buildInstallPlan(false, -1, "", t.TempDir(), nil, true); err == nil {
		t.Fatal("path without skill succeeded, want error")
	}
	if _, err := buildInstallPlan(false, -1, "all", t.TempDir(), nil, true); err == nil {
		t.Fatal("all with path succeeded, want error")
	}
	if _, err := buildInstallPlan(false, -1, "", "", []string{"123"}, true); err == nil {
		t.Fatal("legacy revision without chromium flag succeeded, want error")
	}
	if _, err := buildInstallPlan(true, -1, "", "", []string{"abc"}, true); err == nil {
		t.Fatal("invalid legacy revision succeeded, want error")
	}
	if _, err := buildInstallPlan(true, 123, "", "", []string{"456"}, true); err == nil {
		t.Fatal("extra arg with explicit revision succeeded, want error")
	}
}
