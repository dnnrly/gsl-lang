package gsl_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnnrly/gsl-lang"
)

const skillDir = ".agents/skills/gsl-modelling"

// symlinkedDocs are the repository documents exposed inside the skill via
// relative symlinks. They are validated upstream (TestMarkdownCodeBlocks),
// so this test only asserts the links resolve.
var symlinkedDocs = map[string]string{
	"references/GSL_GUIDE.md":          "GSL_GUIDE.md",
	"references/GQL_GUIDE.md":          "GQL_GUIDE.md",
	"references/GRAMMAR.md":            "GRAMMAR.md",
	"references/QUERY_GRAMMAR.md":      "QUERY_GRAMMAR.md",
	"references/modelling-with-gsl.md": "docs/tutorials/modelling-with-gsl.md",
}

// TestAgentSkillStructure validates the skill's packaging contract:
// frontmatter name matches the directory, a concise description with no
// XML-style tags is present, a (CC) license is declared, and the symlinked
// reference docs resolve to real repository files.
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

	for link, target := range symlinkedDocs {
		linkPath := filepath.Join(skillDir, link)
		fi, err := os.Lstat(linkPath)
		if err != nil {
			t.Errorf("expected symlink %s: %v", linkPath, err)
			continue
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is not a symlink (mode %v)", linkPath, fi.Mode())
			continue
		}
		if _, err := os.Stat(linkPath); err != nil {
			t.Errorf("symlink %s does not resolve: %v", linkPath, err)
			continue
		}
		if _, err := os.Stat(target); err != nil {
			t.Errorf("symlink %s target %q does not exist", linkPath, target)
		}
	}
}

// TestAgentSkillCodeBlocks validates every gsl/gql/invalid-gsl/invalid-gql
// block in the AUTHORED skill markdown. Symlinked repository docs are
// skipped: they are already validated by TestMarkdownCodeBlocks, and on
// filesystems without symlink support the link target text is not GSL.
func TestAgentSkillCodeBlocks(t *testing.T) {
	var authored []string
	err := filepath.WalkDir(skillDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") && d.Type()&os.ModeSymlink == 0 {
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
