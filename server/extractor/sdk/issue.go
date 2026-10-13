// SPDX-License-Identifier: AGPL-3.0-or-later

package sdk

// IssuePreview is the provider independent JSON payload for the issue template.
// Text fields are plain text. BodyHTML must be sanitized before serialization.
type IssuePreview struct {
	Title      string         `json:"title"`
	Number     string         `json:"number,omitempty"`
	Repository string         `json:"repository,omitempty"`
	URL        string         `json:"url,omitempty"`
	State      string         `json:"state,omitempty"`
	Labels     []string       `json:"labels,omitempty"`
	Body       IssueComment   `json:"body"`
	Comments   []IssueComment `json:"comments"`
}

// IssueComment represents an issue description or one saved comment.
// CreatedAt may contain a display date when the source has no machine readable date.
type IssueComment struct {
	Author    string `json:"author,omitempty"`
	AuthorURL string `json:"authorURL,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	URL       string `json:"url,omitempty"`
	BodyHTML  string `json:"bodyHTML"`
}
