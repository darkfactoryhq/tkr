package cmd

import (
	"testing"
)

func TestParsePRRef_FullURL(t *testing.T) {
	ghArg, repoFlag := parsePRRef("https://github.com/owner/repo/pull/123")
	if ghArg != "https://github.com/owner/repo/pull/123" {
		t.Errorf("ghArg = %q, want full URL", ghArg)
	}
	if repoFlag != "" {
		t.Errorf("repoFlag = %q, want empty", repoFlag)
	}
}

func TestParsePRRef_ShortHash(t *testing.T) {
	ghArg, repoFlag := parsePRRef("#123")
	if ghArg != "123" {
		t.Errorf("ghArg = %q, want %q", ghArg, "123")
	}
	if repoFlag != "" {
		t.Errorf("repoFlag = %q, want empty", repoFlag)
	}
}

func TestParsePRRef_PlainNumber(t *testing.T) {
	ghArg, repoFlag := parsePRRef("123")
	if ghArg != "123" {
		t.Errorf("ghArg = %q, want %q", ghArg, "123")
	}
	if repoFlag != "" {
		t.Errorf("repoFlag = %q, want empty", repoFlag)
	}
}

func TestParsePRRef_RepoQualified(t *testing.T) {
	ghArg, repoFlag := parsePRRef("owner/repo#123")
	if ghArg != "123" {
		t.Errorf("ghArg = %q, want %q", ghArg, "123")
	}
	if repoFlag != "owner/repo" {
		t.Errorf("repoFlag = %q, want %q", repoFlag, "owner/repo")
	}
}

func TestParsePRRef_Whitespace(t *testing.T) {
	ghArg, repoFlag := parsePRRef("  #456  ")
	if ghArg != "456" {
		t.Errorf("ghArg = %q, want %q", ghArg, "456")
	}
	if repoFlag != "" {
		t.Errorf("repoFlag = %q, want empty", repoFlag)
	}
}

func TestParsePRRef_HTTPUrl(t *testing.T) {
	ghArg, repoFlag := parsePRRef("http://github.com/owner/repo/pull/99")
	if ghArg != "http://github.com/owner/repo/pull/99" {
		t.Errorf("ghArg = %q, want full URL", ghArg)
	}
	if repoFlag != "" {
		t.Errorf("repoFlag = %q, want empty", repoFlag)
	}
}
