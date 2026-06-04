package bootstrap

import (
	"regexp"
	"strings"
	"testing"
)

// forbiddenPatterns describe content that must never ship inside a publicly
// distributed skill: client/company names, secrets, contact addresses, and
// internal network references. These are intentionally generic (no secret
// deny-list is embedded here, since that list would itself be sensitive) yet
// strong enough to catch the class of leak we have seen in practice
// (e.g. a "<client>-corp" project reference).
var forbiddenPatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"company/client name (-corp suffix)", regexp.MustCompile(`(?i)\b[a-z][a-z0-9]*-corp\b`)},
	{"company marker (株式会社/有限会社/(株))", regexp.MustCompile(`株式会社|有限会社|（株）|\(株\)`)},
	{"secret/API key (hex 32+)", regexp.MustCompile(`\b[0-9a-fA-F]{32,}\b`)},
	{"email address", regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)},
	{"private IP", regexp.MustCompile(`\b(?:10\.\d{1,3}|192\.168|172\.(?:1[6-9]|2\d|3[01]))\.\d`)},
	{"internal hostname (.local/.internal)", regexp.MustCompile(`\b[a-z0-9][a-z0-9\-]*\.(?:local|internal)\b`)},
}

// allowedSubstrings are known-safe fictional placeholders that may legitimately
// match a pattern above. Add entries here (never real client data) if a
// genuine documentation example trips the scanner.
var allowedSubstrings = []string{
	"your-pleasanter.example.com",
	"pleasanter.example.com",
	"example.com",
}

func isAllowed(match string) bool {
	for _, a := range allowedSubstrings {
		if strings.Contains(match, a) {
			return true
		}
	}
	return false
}

func scanForbidden(text string) []string {
	var hits []string
	for _, p := range forbiddenPatterns {
		for _, m := range p.re.FindAllString(text, -1) {
			if isAllowed(m) {
				continue
			}
			hits = append(hits, p.name+": "+m)
		}
	}
	return hits
}

// TestEmbeddedSkills_NoForbiddenContent gates every build: a bundled skill that
// contains a client/company name, secret, email, or internal host fails CI
// before it can be published.
func TestEmbeddedSkills_NoForbiddenContent(t *testing.T) {
	skills, err := Skills()
	if err != nil {
		t.Fatalf("Skills(): %v", err)
	}
	if len(skills) == 0 {
		t.Fatal("no embedded skills found")
	}
	for _, s := range skills {
		for _, hit := range scanForbidden(s.Raw) {
			t.Errorf("skill %q contains forbidden content [%s] — replace with a fictional placeholder", s.Name, hit)
		}
	}

	// Agents, commands, and rules are published for AgentClaude, so they are
	// subject to the same scan.
	for _, group := range []struct {
		kind string
		load func() ([]Doc, error)
	}{
		{"agent", Agents},
		{"command", Commands},
		{"rule", Rules},
	} {
		docs, err := group.load()
		if err != nil {
			t.Fatalf("%s load: %v", group.kind, err)
		}
		for _, d := range docs {
			for _, hit := range scanForbidden(d.Raw) {
				t.Errorf("%s %q contains forbidden content [%s] — replace with a fictional placeholder", group.kind, d.Name, hit)
			}
		}
	}
}

// TestForbiddenPatterns_CatchKnownBad proves the scanner actually fires, so it
// cannot silently rot into a no-op that passes everything.
func TestForbiddenPatterns_CatchKnownBad(t *testing.T) {
	cases := map[string]string{
		"client name":      "このパターンは acme-corp 案件の scripts で実装",
		"company marker":   "発注先は 株式会社サンプル です",
		"api key":          "api_key: 0123456789abcdef0123456789abcdef0123",
		"email":            "問い合わせは support@example.co.jp まで",
		"private ip":       "サーバーは 172.29.224.1 で稼働",
		"internal host":    "エンドポイントは pleasanter.internal",
	}
	for label, sample := range cases {
		if hits := scanForbidden(sample); len(hits) == 0 {
			t.Errorf("%s: scanner failed to flag %q", label, sample)
		}
	}
}

// TestForbiddenPatterns_AllowFictional ensures documentation placeholders do
// not trip the scanner.
func TestForbiddenPatterns_AllowFictional(t *testing.T) {
	ok := []string{
		"plsnt config set --url https://your-pleasanter.example.com",
		"SiteId 100 のレコードを取得",
		"http://localhost/mcp で接続",
	}
	for _, s := range ok {
		if hits := scanForbidden(s); len(hits) > 0 {
			t.Errorf("false positive on safe text %q: %v", s, hits)
		}
	}
}
