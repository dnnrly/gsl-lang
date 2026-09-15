package gsl_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
)

const skillDir = ".agents/skills/gsl-modelling"

// mirroredDocs are the repository documents carried inside the skill.
// They must be dereferenced copies (NOT symlinks): relative symlinks do
// not survive skill packaging, Windows checkouts, or web viewers. Each
// mirror is maintained upstream - TestAgentSkillStructure enforces that
// every mirror still matches its upstream file once the banner is
// stripped, so the copies cannot silently drift.
var mirroredDocs = map[string]string{
	"references/GSL_GUIDE.md":          "GSL_GUIDE.md",
	"references/GQL_GUIDE.md":          "GQL_GUIDE.md",
	"references/GRAMMAR.md":            "GRAMMAR.md",
	"references/QUERY_GRAMMAR.md":      "QUERY_GRAMMAR.md",
	"references/modelling-with-gsl.md": "docs/tutorials/modelling-with-gsl.md",
}

// mirrorBanner is the maintenance banner prepended to each mirror.
var mirrorBanner = regexp.MustCompile(`(?s)<!--\nMaintained upstream at .*?-->`)

// TestAgentSkillStructure validates the skill's packaging contract:
// frontmatter name matches the directory, a concise description with no
// XML-style tags is present, a (CC) license is declared, and the mirrored
// reference docs are regular files that still match their upstream source.
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
		got := strings.TrimSpace(mirrorBanner.ReplaceAllString(string(mirrorContent), ""))
		want := strings.TrimSpace(string(upstreamContent))
		if got != want {
			t.Errorf("mirror %s has drifted from upstream %s; refresh it from the repository", mirror, upstream)
		}
	}
}

// TestAgentSkillCodeBlocks validates every gsl/gql/invalid-gsl/invalid-gql
// block in the skill markdown, mirrors included. The mirrors re-validate
// the same blocks TestMarkdownCodeBlocks covers upstream, so any drift is
// caught in both places.
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
