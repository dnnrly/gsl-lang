package gsl_test

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
)

const skillDir = ".agents/skills/gsl-modelling"

// mirroredDocs are the repository documents carried inside the skill as plain
// copies (NOT symlinks): relative symlinks do not survive skill packaging,
// Windows checkouts, or web viewers, and a dangling symlink makes `cp -r`
// refuse the whole copy. Each mirror is a copy of its upstream file with
// relative links re-resolved from references/; TestAgentSkillMirrorDrift keeps
// them in sync, so the copies cannot silently drift.
var mirroredDocs = map[string]string{
	"references/GSL_GUIDE.md":          "GSL_GUIDE.md",
	"references/GQL_GUIDE.md":          "GQL_GUIDE.md",
	"references/GRAMMAR.md":            "GRAMMAR.md",
	"references/QUERY_GRAMMAR.md":      "QUERY_GRAMMAR.md",
	"references/modelling-with-gsl.md": "docs/tutorials/modelling-with-gsl.md",
}

// markdownLinkRe matches markdown inline link destinations.
var markdownLinkRe = regexp.MustCompile(`\]\(([^)]+)\)`)

// rewriteRelativeLinks re-resolves markdown link targets that reference
// repository files so they still resolve from the mirror's directory after
// the upstream document is copied into references/. Absolute URLs, anchor-only
// links and fragment-carrying targets are left untouched; any trailing #anchor
// is preserved on relocated targets.
func rewriteRelativeLinks(mirror, upstream, content string) string {
	mirrorDir := filepath.Dir(filepath.FromSlash(mirror))
	upstreamDir := filepath.Dir(filepath.FromSlash(upstream))
	return markdownLinkRe.ReplaceAllStringFunc(content, func(whole string) string {
		target := whole[2 : len(whole)-1]
		path, anchor := target, ""
		if i := strings.Index(path, "#"); i >= 0 {
			path, anchor = path[:i], path[i:]
		}
		if path == "" ||
			strings.HasPrefix(path, "http://") ||
			strings.HasPrefix(path, "https://") ||
			strings.HasPrefix(path, "mailto:") ||
			strings.HasPrefix(path, "/") {
			return whole
		}
		projectPath := filepath.Join(upstreamDir, filepath.FromSlash(path))
		fixed, err := filepath.Rel(mirrorDir, projectPath)
		if err != nil {
			return whole
		}
		return "](" + filepath.ToSlash(fixed) + anchor + ")"
	})
}

// updateMirrors makes TestAgentSkillMirrorDrift rewrite the reference mirrors
// from their upstream sources instead of failing on drift.
var updateMirrors = flag.Bool("update", false, "regenerate the reference mirrors from their upstream sources")

// TestAgentSkillStructure validates the skill's packaging contract:
// frontmatter name matches the directory, a concise description with no
// XML-style tags is present, and a (CC) license is declared.
func TestAgentSkillStructure(t *testing.T) {
	skillFile := filepath.Join(skillDir, "SKILL.md")
	frontmatter, err := readFrontmatter(skillFile)
	if err != nil {
		t.Fatalf("SKILL.md frontmatter: %v", err)
	}

	if got := frontmatter["name"]; got != "gsl-modelling" {
		t.Errorf("frontmatter name = %q, want %q (must match the skill directory)", got, "gsl-modelling")
	}

	desc, ok := frontmatter["description"]
	if !ok || strings.TrimSpace(desc) == "" {
		t.Errorf("SKILL.md description is missing or empty")
	} else {
		if len([]rune(desc)) >= 1024 {
			t.Errorf("SKILL.md description is %d chars; keep it under 1024", len([]rune(desc)))
		}
		if strings.ContainsAny(desc, "<>") {
			t.Errorf("SKILL.md description must not contain XML-style tags: %q", desc)
		}
	}

	if _, ok := frontmatter["license"]; !ok {
		t.Errorf("SKILL.md frontmatter is missing a license field")
	}
}

// TestAgentSkillMirrorDrift prevents the reference mirrors from silently
// drifting from the authoritative repository documents. Each mirror must be a
// regular file (not a symlink) whose content equals its upstream source once
// relative links are re-resolved from references/. Run with -update to
// rewrite the mirrors from the current upstream sources.
func TestAgentSkillMirrorDrift(t *testing.T) {
	refreshed := 0
	for mirror, upstream := range mirroredDocs {
		mirrorPath := filepath.Join(skillDir, mirror)
		fi, err := os.Lstat(mirrorPath)
		if err != nil {
			t.Errorf("missing mirror %s: %v", mirrorPath, err)
			continue
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Errorf("%s is a symlink; mirrors must be regular files (symlinks do not survive packaging)", mirrorPath)
			continue
		}
		mirrorContent, err := os.ReadFile(mirrorPath)
		if err != nil {
			t.Errorf("failed to read %s: %v", mirrorPath, err)
			continue
		}
		upstreamContent, err := os.ReadFile(upstream)
		if err != nil {
			t.Errorf("failed to read upstream %s: %v", upstream, err)
			continue
		}
		want := rewriteRelativeLinks(mirrorPath, upstream, string(upstreamContent))
		if strings.TrimSpace(string(mirrorContent)) == strings.TrimSpace(want) {
			continue
		}
		if *updateMirrors {
			if err := os.WriteFile(mirrorPath, []byte(want), 0o644); err != nil {
				t.Errorf("failed to refresh %s: %v", mirrorPath, err)
				continue
			}
			refreshed++
			t.Logf("refreshed %s from %s", mirror, upstream)
			continue
		}
		t.Errorf("mirror %s has drifted from upstream %s; refresh with `go test -run TestAgentSkillMirrorDrift -update .`", mirror, upstream)
	}
	if *updateMirrors {
		if refreshed == 0 {
			t.Log("mirrors are already up to date")
		} else {
			t.Logf("refreshed %d mirror(s)", refreshed)
		}
	}
}

// TestAgentSkillReferenceLinks verifies that every markdown link in the
// skill (mirrors included) that points at a repository file resolves to a
// real file. Anchor-only links and absolute URLs are exempt.
func TestAgentSkillReferenceLinks(t *testing.T) {
	var mdFiles []string
	err := filepath.WalkDir(skillDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk %s: %v", skillDir, err)
	}

	for _, mdFile := range mdFiles {
		content, err := os.ReadFile(mdFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", mdFile, err)
			continue
		}
		for i, target := range extractLinkTargets(string(content)) {
			path, _, _ := strings.Cut(target, "#")
			if path == "" ||
				strings.HasPrefix(path, "http://") ||
				strings.HasPrefix(path, "https://") ||
				strings.HasPrefix(path, "mailto:") ||
				strings.HasPrefix(path, "/") {
				continue
			}
			resolved := filepath.Join(filepath.Dir(mdFile), filepath.FromSlash(path))
			if _, err := os.Stat(resolved); err != nil {
				t.Errorf("%s link %d resolves to %s: %v", mdFile, i+1, resolved, err)
			}
		}
	}
}

// extractLinkTargets returns every markdown inline link destination.
func extractLinkTargets(content string) []string {
	var targets []string
	for _, m := range markdownLinkRe.FindAllStringSubmatch(content, -1) {
		targets = append(targets, m[1])
	}
	return targets
}

// TestAgentSkillCodeBlocks validates every gsl/gql/invalid-gsl/invalid-gql
// block in the skill markdown, mirror copies included. The mirrors re-validate
// the same blocks TestMarkdownCodeBlocks covers upstream, so any copied
// content that fails to parse is caught in both places.
func TestAgentSkillCodeBlocks(t *testing.T) {
	var authored []string
	err := filepath.WalkDir(skillDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			authored = append(authored, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk %s: %v", skillDir, err)
	}

	if len(authored) == 0 {
		t.Fatalf("no authored markdown files found under %s", skillDir)
	}

	validGSL, invalidGSL, validGQL, invalidGQL := 0, 0, 0, 0
	for _, mdFile := range authored {
		content, err := os.ReadFile(mdFile)
		if err != nil {
			t.Errorf("failed to read %s: %v", mdFile, err)
			continue
		}
		for i, block := range extractCodeBlocks(string(content)) {
			switch block.language {
			case "gsl":
				validGSL++
				if err := testValidGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v\nCode:\n%s", mdFile, i+1, block.lineNumber, err, block.code)
				}
			case "invalid-gsl":
				invalidGSL++
				if err := testInvalidGSL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): expected parse to fail, but got: %v\nCode:\n%s", mdFile, i+1, block.lineNumber, err, block.code)
				}
			case "gql":
				validGQL++
				if err := testValidGQL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): %v\nCode:\n%s", mdFile, i+1, block.lineNumber, err, block.code)
				}
			case "invalid-gql":
				invalidGQL++
				if err := testInvalidGQL(block.code); err != nil {
					t.Errorf("%s block %d (line %d): expected parse to fail, but got: %v\nCode:\n%s", mdFile, i+1, block.lineNumber, err, block.code)
				}
			}
		}
	}
	t.Logf("Skill authored blocks: %d gsl, %d invalid-gsl, %d gql, %d invalid-gql", validGSL, invalidGSL, validGQL, invalidGQL)
	if validGSL == 0 && invalidGSL == 0 && validGQL == 0 && invalidGQL == 0 {
		t.Fatalf("no gsl/gql code blocks found in authored skill markdown")
	}
}

// TestAgentSkillScriptSyntax checks the optional validate.sh script compiles.
func TestAgentSkillScriptSyntax(t *testing.T) {
	path := filepath.Join(skillDir, "scripts", "validate.sh")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("validate.sh missing: %v", err)
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	out, err := exec.Command("bash", "-n", path).CombinedOutput()
	if err != nil {
		t.Errorf("bash -n %s failed: %v\n%s", path, err, out)
	}
}

// TestSkillBehaviourOracleParses guarantees the misleading-source oracle
// (examples/skill-behaviour/expected-fidelity.gsl) stays structurally valid.
func TestSkillBehaviourOracleParses(t *testing.T) {
	for _, fixture := range []string{
		"examples/skill-behaviour/misleading-source.md",
		"examples/skill-behaviour/README.md",
	} {
		if _, err := os.Stat(fixture); err != nil {
			t.Errorf("behavioural fixture missing: %s", fixture)
		}
	}

	oracle := "examples/skill-behaviour/expected-fidelity.gsl"
	content, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatalf("failed to read %s: %v", oracle, err)
	}
	_, parseErr := gsl.Parse(strings.NewReader(string(content)))
	if parseErr != nil && parseErr.HasError() {
		t.Fatalf("%s failed to parse: %v", oracle, parseErr)
	}
}

// readFrontmatter extracts the leading YAML frontmatter block of a markdown
// file (--- delimited) as a string map. Indented continuation lines of a
// block scalar are joined into the preceding key with a single space.
func readFrontmatter(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(content)
	const open = "---\n"
	if !strings.HasPrefix(text, open) {
		return nil, os.ErrInvalid
	}
	rest := text[len(open):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, os.ErrInvalid
	}

	fields := map[string]string{}
	current := ""
	for _, line := range strings.Split(rest[:end], "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			if current != "" {
				fields[current] = strings.TrimSpace(fields[current] + " " + strings.TrimSpace(line))
			}
			continue
		}
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" || value == ">" || value == ">-" || value == "|" || value == "|-" {
			current = key
			fields[key] = ""
			continue
		}
		fields[key] = strings.Trim(value, "\"'")
		current = key
	}
	return fields, nil
}
