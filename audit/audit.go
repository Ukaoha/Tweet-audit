package audit

import (
	"context"
	"fmt"
	"strings"

	"tweet-audit/archive"
	"tweet-audit/gemini"
)

type Auditor struct {
	client   *gemini.Client
	criteria Criteria
	username string
}

func NewAuditor(client *gemini.Client, criteria Criteria, username string) *Auditor {
	return &Auditor{
		client:   client,
		criteria: criteria,
		username: username,
	}
}

func (a *Auditor) Audit(ctx context.Context, tweet archive.Tweet) (Verdict, error) {
	verdict := Verdict{
		TweetID:   tweet.IDStr,
		Username:  a.username,
		URL:       fmt.Sprintf("https://x.com/%s/status/%s", a.username, tweet.IDStr),
		Text:      tweet.FullText,
		CreatedAt: tweet.CreatedAt,
	}

	if !a.criteria.IncludeRetweets && tweet.Retweeted {
		verdict.Flag = false
		return verdict, nil
	}

	if a.criteria.MinFavoritesToKeep > 0 || a.criteria.MinRetweetsToKeep > 0 {
		favCount, retCount := parseInt(tweet.FavoriteCount), parseInt(tweet.RetweetCount)
		if favCount >= a.criteria.MinFavoritesToKeep && retCount >= a.criteria.MinRetweetsToKeep {
			verdict.Flag = false
			return verdict, nil
		}
	}

	prompt := a.buildPrompt(tweet)
	response, err := a.client.Generate(ctx, prompt)
	if err != nil {
		return verdict, fmt.Errorf("gemini evaluation failed: %w", err)
	}

	verdict.Flag, verdict.Reason = a.parseResponse(response)
	return verdict, nil
}

func (a *Auditor) buildPrompt(tweet archive.Tweet) string {
	criteriaStr := a.formatCriteria()

	prompt := fmt.Sprintf(`You are a Twitter content auditor. Evaluate the following tweet against these criteria:

%s

Tweet to evaluate:
---
%s
---

Respond with exactly one line in this format:
FLAG: <yes|no> | REASON: <brief reason or "no issues">

Be concise and objective.`, criteriaStr, tweet.FullText)

	return prompt
}

func (a *Auditor) formatCriteria() string {
	var parts []string

	if len(a.criteria.ForbiddenWords) > 0 {
		parts = append(parts, fmt.Sprintf("- Forbidden words/phrases: %s", strings.Join(a.criteria.ForbiddenWords, ", ")))
	}

	if a.criteria.ProfessionalCheck {
		parts = append(parts, "- Should maintain professional tone and language")
	}

	if a.criteria.Tone != "" {
		parts = append(parts, fmt.Sprintf("- Expected tone: %s", a.criteria.Tone))
	}

	if a.criteria.ExcludePolitics {
		parts = append(parts, "- Flag political content or political commentary")
	}

	if len(a.criteria.CustomRules) > 0 {
		parts = append(parts, fmt.Sprintf("- Custom rules: %s", strings.Join(a.criteria.CustomRules, "; ")))
	}

	if len(parts) == 0 {
		parts = append(parts, "- No specific criteria defined")
	}

	return strings.Join(parts, "\n")
}

func (a *Auditor) parseResponse(response string) (bool, string) {
	response = strings.TrimSpace(response)

	flag := strings.Contains(strings.ToLower(response), "flag: yes")

	reason := response
	if idx := strings.Index(response, "REASON:"); idx != -1 {
		reason = strings.TrimSpace(response[idx+7:])
	}

	return flag, reason
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
