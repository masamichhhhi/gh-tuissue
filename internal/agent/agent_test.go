package agent

import (
	"strings"
	"testing"

	"github.com/masamichhhhi/gh-tuissue/internal/config"
	"github.com/masamichhhhi/gh-tuissue/internal/domain"
)

func TestCommand_DefaultArgs(t *testing.T) {
	issue := domain.Issue{Number: 548, Title: "Directorio S.A. falla"}
	got := Command(config.AgentAction{Key: "a", Skill: "ticket_resolve"}, issue)
	want := []string{"claude", "--bg", "--permission-mode", "auto", "-n", "548-directorio-s-a-falla", "/ticket_resolve 548"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("Command() = %q, want %q", got, want)
	}
}

func TestCommand_CustomArgsAndLeadingSlash(t *testing.T) {
	issue := domain.Issue{Number: 7, Title: "Bug", URL: "https://github.com/o/r/issues/7"}
	got := Command(config.AgentAction{Key: "b", Skill: "/bugfix", Args: "{{url}} {{title}}"}, issue)
	if got[len(got)-1] != "/bugfix https://github.com/o/r/issues/7 Bug" {
		t.Errorf("prompt = %q", got[len(got)-1])
	}
}

func TestSessionName_Truncates(t *testing.T) {
	issue := domain.Issue{Number: 123, Title: strings.Repeat("palabra ", 20)}
	name := SessionName(issue)
	if len(name) > maxNameLen {
		t.Errorf("len(%q) = %d, want <= %d", name, len(name), maxNameLen)
	}
	if strings.HasSuffix(name, "-") {
		t.Errorf("name %q ends with dash", name)
	}
	if !strings.HasPrefix(name, "123-palabra") {
		t.Errorf("name = %q", name)
	}
}

func TestSessionName_EmptyTitle(t *testing.T) {
	if got := SessionName(domain.Issue{Number: 9}); got != "9" {
		t.Errorf("SessionName = %q, want %q", got, "9")
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello, World!":       "hello-world",
		"  spaces   around  ": "spaces-around",
		"Ñandú café 2024":     "and-caf-2024",
		"---":                 "",
		"CamelCase_and.dots":  "camelcase-and-dots",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseOutput(t *testing.T) {
	out := "backgrounded · \x1b[36m25e6d329\x1b[39m · bgtest-42\n" +
		"\x1b[2m  claude agents             list sessions\x1b[22m\n"
	l, err := ParseOutput(out)
	if err != nil {
		t.Fatalf("ParseOutput error: %v", err)
	}
	if l.ID != "25e6d329" || l.Name != "bgtest-42" {
		t.Errorf("ParseOutput = %+v", l)
	}
}

func TestParseOutput_Unexpected(t *testing.T) {
	_, err := ParseOutput("\x1b[31mError: something broke\x1b[0m\nmore")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Error: something broke") {
		t.Errorf("error = %v, want first line of output", err)
	}
}

func TestRun_MissingBinary(t *testing.T) {
	_, err := Run(t.TempDir(), []string{"gh-tuissue-definitely-missing-binary", "--bg"})
	if err == nil || !strings.Contains(err.Error(), "not found in PATH") {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestValidate(t *testing.T) {
	in := []config.AgentAction{
		{Key: "a", Skill: "ticket_resolve"},
		{Key: "", Skill: "x"},
		{Key: "ab", Skill: "y"},
		{Key: "b", Skill: ""},
		{Key: "a", Skill: "dup"},
		{Key: "m", Skill: "mail_resolve"},
	}
	valid, warnings := Validate(in)
	if len(valid) != 2 || valid[0].Skill != "ticket_resolve" || valid[1].Skill != "mail_resolve" {
		t.Errorf("valid = %+v", valid)
	}
	if len(warnings) != 4 {
		t.Errorf("warnings = %q, want 4", warnings)
	}
}
