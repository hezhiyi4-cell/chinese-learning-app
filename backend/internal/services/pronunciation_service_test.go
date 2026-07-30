package services

import (
	"strings"
	"testing"
)

func TestBuildPronunciationScriptForInitialTable(t *testing.T) {
	script := buildPronunciationScript("第1课：拼音声母表", "b, p, m, f, d, t, n, l, g, k, h, j, q, x, zh, ch, sh, r, z, c, s")
	if !strings.Contains(script.spokenText, "玻") || !strings.Contains(script.spokenText, "诗") {
		t.Fatalf("expected initial pronunciation script, got %q", script.spokenText)
	}
}

func TestBuildPronunciationScriptForFinalTable(t *testing.T) {
	script := buildPronunciationScript("第2课：拼音韵母表", "a, o, e, i, u, ü, ai, ei, ui, ao, ou, iu, ie, üe, er, an, en, in, un, ün, ang, eng, ing, ong")
	if !strings.Contains(script.spokenText, "迂") || !strings.Contains(script.spokenText, "翁") {
		t.Fatalf("expected final pronunciation script, got %q", script.spokenText)
	}
}

func TestNormalizePinyinToken(t *testing.T) {
	cases := map[string]string{
		"nǐ":     "ni3",
		"hǎo":    "hao3",
		"lüè":    "lve4",
		"zhong1": "zhong1",
	}

	for input, want := range cases {
		got, ok := normalizePinyinToken(input)
		if !ok {
			t.Fatalf("expected token %q to be valid", input)
		}
		if got != want {
			t.Fatalf("normalizePinyinToken(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildPinyinCSSML(t *testing.T) {
	script := buildPronunciationScript("问候", "nǐ hǎo")
	if !strings.Contains(script.spokenText, `<phoneme lang="zh-cn">ni3</phoneme>`) {
		t.Fatalf("expected ni3 phoneme, got %q", script.spokenText)
	}
	if !strings.Contains(script.spokenText, `<phoneme lang="zh-cn">hao3</phoneme>`) {
		t.Fatalf("expected hao3 phoneme, got %q", script.spokenText)
	}
}
