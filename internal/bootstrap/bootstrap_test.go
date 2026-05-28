package bootstrap

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	coreSkillCount   = 11
	coreAgentCount   = 5
	coreCommandCount = 5
	coreRuleCount    = 1
	coreClaudeFiles  = coreSkillCount + coreAgentCount + coreCommandCount + coreRuleCount
)

func TestSkills_ReturnsAllCore(t *testing.T) {
	skills, err := Skills()
	if err != nil {
		t.Fatalf("Skills() error: %v", err)
	}
	if len(skills) != coreSkillCount {
		t.Fatalf("expected %d skills, got %d", coreSkillCount, len(skills))
	}

	// Sorted by name.
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("skills not sorted by name: %v", names)
	}

	for _, s := range skills {
		if s.Name == "" {
			t.Errorf("skill has empty Name")
		}
		if s.Description == "" {
			t.Errorf("skill %q has empty Description", s.Name)
		}
		if strings.TrimSpace(s.Body) == "" {
			t.Errorf("skill %q has empty Body", s.Name)
		}
		// Body must not contain the frontmatter delimiter region.
		if strings.HasPrefix(strings.TrimSpace(s.Body), "---") {
			t.Errorf("skill %q Body still contains frontmatter", s.Name)
		}
		if !strings.Contains(s.Raw, "name: "+s.Name) {
			t.Errorf("skill %q Raw missing name frontmatter", s.Name)
		}
	}
}

func TestSkillNames_IncludesGuide(t *testing.T) {
	names, err := SkillNames()
	if err != nil {
		t.Fatalf("SkillNames() error: %v", err)
	}
	if len(names) != coreSkillCount {
		t.Fatalf("expected %d names, got %d", coreSkillCount, len(names))
	}
	if !contains(names, "plsnt-guide") {
		t.Errorf("expected plsnt-guide in names: %v", names)
	}
	if !contains(names, "troubleshooting") {
		t.Errorf("expected troubleshooting in names: %v", names)
	}
}

func TestParseAgent(t *testing.T) {
	cases := map[string]Agent{
		"claude":  AgentClaude,
		"codex":   AgentCodex,
		"gemini":  AgentGemini,
		"generic": AgentGeneric,
	}
	for in, want := range cases {
		got, err := ParseAgent(in)
		if err != nil {
			t.Errorf("ParseAgent(%q) error: %v", in, err)
		}
		if got != want {
			t.Errorf("ParseAgent(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := ParseAgent("cursor-nope"); err == nil {
		t.Errorf("ParseAgent(invalid) expected error, got nil")
	}
}

func TestInstall_Claude(t *testing.T) {
	dir := t.TempDir()
	written, err := Install(dir, AgentClaude)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if len(written) != coreClaudeFiles {
		t.Fatalf("expected %d files written, got %d", coreClaudeFiles, len(written))
	}

	// Each skill is at .claude/skills/<name>/SKILL.md
	guide := filepath.Join(dir, ".claude", "skills", "plsnt-guide", "SKILL.md")
	data, err := os.ReadFile(guide)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", guide, err)
	}
	if !strings.Contains(string(data), "name: plsnt-guide") {
		t.Errorf("plsnt-guide SKILL.md missing frontmatter")
	}

	// Agents, commands, and rules are flat .md files under .claude/.
	for _, p := range []string{
		filepath.Join(dir, ".claude", "agents", "pleasanter-api-expert.md"),
		filepath.Join(dir, ".claude", "commands", "site-build.md"),
		filepath.Join(dir, ".claude", "rules", "domain-glossary.md"),
	} {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Errorf("expected %s to exist: %v", p, statErr)
		}
	}
}

func TestAgentsCommandsRules_Counts(t *testing.T) {
	agents, err := Agents()
	if err != nil || len(agents) != coreAgentCount {
		t.Errorf("Agents() = %d (err %v), want %d", len(agents), err, coreAgentCount)
	}
	cmds, err := Commands()
	if err != nil || len(cmds) != coreCommandCount {
		t.Errorf("Commands() = %d (err %v), want %d", len(cmds), err, coreCommandCount)
	}
	rules, err := Rules()
	if err != nil || len(rules) != coreRuleCount {
		t.Errorf("Rules() = %d (err %v), want %d", len(rules), err, coreRuleCount)
	}
	if len(agents) > 0 && !contains(docNames(agents), "pleasanter-api-expert") {
		t.Errorf("expected pleasanter-api-expert agent: %v", docNames(agents))
	}
	if len(rules) > 0 && rules[0].Name != "domain-glossary" {
		t.Errorf("expected domain-glossary rule, got %q", rules[0].Name)
	}
}

func docNames(docs []Doc) []string {
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.Name
	}
	return out
}

func TestInstall_NonClaude_FoldsRulesButNotAgentsCommands(t *testing.T) {
	dir := t.TempDir()
	written, err := Install(dir, AgentCodex)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	// Only AGENTS.md is written — no .claude/ tree for non-Claude agents.
	if len(written) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(written), written)
	}
	if _, statErr := os.Stat(filepath.Join(dir, ".claude")); statErr == nil {
		t.Errorf(".claude/ must not be created for codex agent")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	content := string(data)
	// Rule content (domain glossary) is folded in.
	if !strings.Contains(content, "rule: domain-glossary") || !strings.Contains(content, "SiteID") {
		t.Errorf("AGENTS.md should fold in the domain-glossary rule")
	}
	// Sub-agent / command names must NOT be folded in (they don't function).
	if strings.Contains(content, "Pleasanter API Expert") {
		t.Errorf("AGENTS.md should not contain sub-agent content")
	}
}

func TestInstall_Codex_WritesAgentsMD(t *testing.T) {
	dir := t.TempDir()
	written, err := Install(dir, AgentCodex)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 file written, got %d: %v", len(written), written)
	}
	agentsPath := filepath.Join(dir, "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("expected AGENTS.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, beginMarker) || !strings.Contains(content, endMarker) {
		t.Errorf("AGENTS.md missing markers")
	}
	// All skill names must appear as sections.
	names, _ := SkillNames()
	for _, n := range names {
		if !strings.Contains(content, "## "+n) {
			t.Errorf("AGENTS.md missing section for %q", n)
		}
	}
}

func TestInstall_Gemini_WritesGeminiMD(t *testing.T) {
	dir := t.TempDir()
	if _, err := Install(dir, AgentGemini); err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "GEMINI.md")); err != nil {
		t.Fatalf("expected GEMINI.md: %v", err)
	}
}

func TestInstall_Generic_WritesAgentsMD(t *testing.T) {
	dir := t.TempDir()
	if _, err := Install(dir, AgentGeneric); err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md: %v", err)
	}
}

func TestInstall_Codex_MergeIsIdempotentAndPreservesUserContent(t *testing.T) {
	dir := t.TempDir()
	agentsPath := filepath.Join(dir, "AGENTS.md")
	userContent := "# My project rules\n\nDo not touch the database.\n"
	if err := os.WriteFile(agentsPath, []byte(userContent), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Install(dir, AgentCodex); err != nil {
		t.Fatalf("first Install error: %v", err)
	}
	if _, err := Install(dir, AgentCodex); err != nil {
		t.Fatalf("second Install error: %v", err)
	}

	data, _ := os.ReadFile(agentsPath)
	content := string(data)

	// User content preserved.
	if !strings.Contains(content, "Do not touch the database.") {
		t.Errorf("user content lost after install")
	}
	// Exactly one plsnt block (idempotent).
	if got := strings.Count(content, beginMarker); got != 1 {
		t.Errorf("expected exactly 1 begin marker after double install, got %d", got)
	}
	if got := strings.Count(content, endMarker); got != 1 {
		t.Errorf("expected exactly 1 end marker after double install, got %d", got)
	}
}

func contains(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}
