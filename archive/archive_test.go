package archive

import (
	"testing"
)

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "typical twitter js file",
			input:    []byte(`window.YTD.tweets.part0 = [{"tweet": {}}]`),
			expected: []byte(` [{"tweet": {}}]`),
		},
		{
			name:     "with spaces around equals",
			input:    []byte(`window.YTD.tweets.part0 = [1, 2, 3]`),
			expected: []byte(` [1, 2, 3]`),
		},
		{
			name:     "no equals sign",
			input:    []byte(`just plain json [{"a": 1}]`),
			expected: []byte(`just plain json [{"a": 1}]`),
		},
		{
			name:     "multiple equals signs (takes first)",
			input:    []byte(`a=b=c`),
			expected: []byte(`b=c`),
		},
		{
			name:     "empty after equals",
			input:    []byte(`prefix=`),
			expected: []byte(``),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripPrefix(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("stripPrefix(%q) = %q, want %q", string(tt.input), string(result), string(tt.expected))
			}
		})
	}
}

func TestTweetFields(t *testing.T) {
	tw := Tweet{
		IDStr:         "123456",
		FullText:      "Hello world",
		CreatedAt:     "Mon Jun 02 07:25:39 +0000 2026",
		FavoriteCount: "42",
		RetweetCount:  "5",
		Lang:          "en",
		Retweeted:     false,
	}

	if tw.IDStr != "123456" {
		t.Errorf("IDStr = %q, want %q", tw.IDStr, "123456")
	}
	if tw.FullText != "Hello world" {
		t.Errorf("FullText = %q, want %q", tw.FullText, "Hello world")
	}
	if tw.FavoriteCount != "42" {
		t.Errorf("FavoriteCount = %q, want %q", tw.FavoriteCount, "42")
	}
	if tw.Retweeted != false {
		t.Errorf("Retweeted = %v, want %v", tw.Retweeted, false)
	}
}

func TestAccountFields(t *testing.T) {
	acc := Account{
		Username:  "testuser",
		AccountID: "1234567890",
	}

	if acc.Username != "testuser" {
		t.Errorf("Username = %q, want %q", acc.Username, "testuser")
	}
	if acc.AccountID != "1234567890" {
		t.Errorf("AccountID = %q, want %q", acc.AccountID, "1234567890")
	}
}
