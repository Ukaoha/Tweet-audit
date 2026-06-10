package archive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func stripPrefix(b []byte) []byte {
	if i := bytes.IndexByte(b, '='); i != -1 {
		return b[i+1:]
	}
	return b
}

// read tweets
func LoadTweets(path string) ([]Tweet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading tweets file: %w", err)
	}

	var wrappers []tweetWrapper
	if err := json.Unmarshal(stripPrefix(raw), &wrappers); err != nil {
		return nil, fmt.Errorf("parsing tweets: %w", err)
	}

	tweets := make([]Tweet, len(wrappers))
	for i, w := range wrappers {
		tweets[i] = w.Tweet
	}
	return tweets, nil
}

// read account
func LoadAccount(path string) (Account, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Account{}, fmt.Errorf("reading account file: %w", err)
	}

	var wrappers []accountWrapper
	if err := json.Unmarshal(stripPrefix(raw), &wrappers); err != nil {
		return Account{}, fmt.Errorf("parsing account: %w", err)
	}
	if len(wrappers) == 0 {
		return Account{}, fmt.Errorf("no account found in %s", path)
	}
	return wrappers[0].Account, nil
}
