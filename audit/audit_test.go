package audit

import (
	"fmt"
	"strings"
	"testing"

	"tweet-audit/archive"
)

func TestFormatCriteria(t *testing.T) {
	tests := []struct {
		name     string
		criteria Criteria
		contains []string
		excludes []string
	}{
		{
			name: "all criteria fields",
			criteria: Criteria{
				ForbiddenWords:    []string{"bad", "worse"},
				ProfessionalCheck: true,
				Tone:              "formal",
				ExcludePolitics:   true,
				CustomRules:       []string{"rule1", "rule2"},
			},
			contains: []string{"bad", "worse", "professional", "formal", "political", "rule1", "rule2"},
			excludes: []string{},
		},
		{
			name: "no criteria",
			criteria: Criteria{
				ForbiddenWords:    []string{},
				ProfessionalCheck: false,
				Tone:              "",
				ExcludePolitics:   false,
				CustomRules:       []string{},
			},
			contains: []string{"No specific criteria"},
			excludes: []string{},
		},
		{
			name: "only forbidden words",
			criteria: Criteria{
				ForbiddenWords:    []string{"spam"},
				ProfessionalCheck: false,
				Tone:              "",
				ExcludePolitics:   false,
				CustomRules:       []string{},
			},
			contains: []string{"spam", "Forbidden"},
			excludes: []string{"professional"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditor := NewAuditor(nil, tt.criteria, "testuser")
			result := auditor.formatCriteria()

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("formatCriteria() missing %q in output:\n%s", s, result)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(result, s) {
					t.Errorf("formatCriteria() should not contain %q in output:\n%s", s, result)
				}
			}
		})
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		wantFlag      bool
		wantReasonHas string
	}{
		{
			name:          "flag yes",
			response:      "FLAG: yes | REASON: contains forbidden word",
			wantFlag:      true,
			wantReasonHas: "forbidden",
		},
		{
			name:          "flag no",
			response:      "FLAG: no | REASON: no issues",
			wantFlag:      false,
			wantReasonHas: "no issues",
		},
		{
			name:          "flag yes lowercase",
			response:      "flag: yes | reason: too casual",
			wantFlag:      true,
			wantReasonHas: "casual",
		},
		{
			name:          "malformed response",
			response:      "random text without FLAG",
			wantFlag:      false,
			wantReasonHas: "random",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditor := NewAuditor(nil, Criteria{}, "testuser")
			flag, reason := auditor.parseResponse(tt.response)

			if flag != tt.wantFlag {
				t.Errorf("parseResponse(%q) flag = %v, want %v", tt.response, flag, tt.wantFlag)
			}
			if !strings.Contains(reason, tt.wantReasonHas) {
				t.Errorf("parseResponse(%q) reason = %q, want to contain %q", tt.response, reason, tt.wantReasonHas)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"123", 123},
		{"-5", -5},
		{"", 0},
		{"invalid", 0},
		{"42extra", 42},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("parseInt(%q)", tt.input), func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAuditorFields(t *testing.T) {
	criteria := Criteria{ForbiddenWords: []string{"test"}}
	auditor := NewAuditor(nil, criteria, "myuser")

	if auditor.username != "myuser" {
		t.Errorf("auditor username = %q, want %q", auditor.username, "myuser")
	}
	if auditor.criteria.ForbiddenWords[0] != "test" {
		t.Errorf("auditor criteria not preserved")
	}
}

func TestBuildPrompt(t *testing.T) {
	criteria := Criteria{
		ForbiddenWords:    []string{"crypto", "NFT"},
		ProfessionalCheck: true,
		Tone:              "respectful",
		ExcludePolitics:   false,
		CustomRules:       []string{"no spam"},
	}

	auditor := NewAuditor(nil, criteria, "testuser")

	tweet := archive.Tweet{
		IDStr:    "123",
		FullText: "Check out this NFT drop",
	}

	prompt := auditor.buildPrompt(tweet)

	if !strings.Contains(prompt, "crypto") {
		t.Error("prompt missing forbidden word 'crypto'")
	}
	if !strings.Contains(prompt, "NFT") {
		t.Error("prompt missing forbidden word 'NFT'")
	}
	if !strings.Contains(prompt, "Check out this NFT drop") {
		t.Error("prompt missing tweet text")
	}
}
