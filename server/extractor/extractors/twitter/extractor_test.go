// SPDX-License-Identifier: AGPL-3.0-or-later

package twitter

import (
	"net/url"
	"strings"
	"testing"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/types"
)

func TestSetConfigRejectsUnknownOptions(t *testing.T) {
	e := &TwitterExtractor{}
	err := e.SetConfig(&config.Extractor{
		Enable:  true,
		Options: map[string]any{"unknown": true},
	})
	if err == nil {
		t.Fatal("SetConfig accepted an unknown option")
	}
}

func TestMatch(t *testing.T) {
	e := &TwitterExtractor{}
	tests := []struct {
		name string
		doc  *document.Document
		want bool
	}{
		{
			name: "X profile",
			doc:  &document.Document{URL: "https://x.com/alice"},
			want: true,
		},
		{
			name: "Twitter feed",
			doc:  &document.Document{URL: "https://mobile.twitter.com/home"},
			want: true,
		},
		{
			name: "extracted tweet",
			doc: &document.Document{
				URL:      "https://example.com/imported",
				Metadata: map[string]any{"type": "tweet"},
			},
			want: true,
		},
		{
			name: "unrelated X subdomain",
			doc:  &document.Document{URL: "https://api.x.com/2/tweets"},
			want: false,
		},
		{
			name: "lookalike host",
			doc:  &document.Document{URL: "https://notx.com/alice/status/123"},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := e.Match(test.doc); got != test.want {
				t.Fatalf("Match() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCanonicalStatusURL(t *testing.T) {
	base, err := url.Parse("https://x.com/home")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		raw        string
		wantURL    string
		wantHandle string
		wantOK     bool
	}{
		{
			name:       "legacy Twitter URL",
			raw:        "https://twitter.com/alice/status/123?s=20#fragment",
			wantURL:    "https://x.com/alice/status/123",
			wantHandle: "alice",
			wantOK:     true,
		},
		{
			name:       "relative X URL",
			raw:        "/Bob_2/status/456/photo/1",
			wantURL:    "https://x.com/Bob_2/status/456",
			wantHandle: "Bob_2",
			wantOK:     true,
		},
		{
			name:    "generic status URL",
			raw:     "https://x.com/i/web/status/789",
			wantURL: "https://x.com/i/status/789",
			wantOK:  true,
		},
		{
			name: "unrelated host",
			raw:  "https://example.com/alice/status/123",
		},
		{
			name: "invalid status identifier",
			raw:  "https://x.com/alice/status/not-a-number",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotURL, gotHandle, gotOK := canonicalStatusURL(test.raw, base)
			if gotURL != test.wantURL || gotHandle != test.wantHandle || gotOK != test.wantOK {
				t.Fatalf(
					"canonicalStatusURL(%q) = (%q, %q, %v), want (%q, %q, %v)",
					test.raw, gotURL, gotHandle, gotOK,
					test.wantURL, test.wantHandle, test.wantOK,
				)
			}
		})
	}
}

func TestExtractSemanticFeed(t *testing.T) {
	d := &document.Document{
		URL:    "https://x.com/githubstatus",
		UserID: 42,
		HTML: `<main itemscope itemtype="https://schema.org/ProfilePage">
			<article data-tweet-id="123" itemscope itemtype="https://schema.org/SocialMediaPosting">
				<meta itemprop="url" content="https://x.com/githubstatus/status/123?ref=feed">
				<meta itemprop="articleBody" content="First status update https://t.co/example">
				<meta itemprop="datePublished" content="2026-08-11T20:06:56.000Z">
				<div hidden itemprop="author" itemscope itemtype="https://schema.org/Person">
					<meta itemprop="name" content="GitHub Status">
					<meta itemprop="alternateName" content="githubstatus">
				</div>
				<img src="https://pbs.twimg.com/media/example.jpg" alt="Incident chart">
			</article>
			<article data-tweet-id="456" itemscope itemtype="https://schema.org/SocialMediaPosting">
				<meta itemprop="url" content="https://twitter.com/githubstatus/status/456">
				<meta itemprop="text" content="Second status update">
				<div hidden itemprop="author" itemscope itemtype="https://schema.org/Person">
					<meta itemprop="name" content="GitHub Status">
					<meta itemprop="alternateName" content="@githubstatus">
				</div>
			</article>
		</main>`,
	}

	state, err := (&TwitterExtractor{}).Extract(d)
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != types.ExtractorStop {
		t.Fatalf("Extract state = %v, want %v", state, types.ExtractorStop)
	}
	if !d.SkipIndexing {
		t.Fatal("source feed was not marked to skip indexing")
	}
	if len(d.ExtraDocuments) != 2 {
		t.Fatalf("ExtraDocuments length = %d, want 2", len(d.ExtraDocuments))
	}

	first := d.ExtraDocuments[0]
	if got, want := first.URL, "https://x.com/githubstatus/status/123"; got != want {
		t.Fatalf("tweet URL = %q, want %q", got, want)
	}
	if got, want := first.Title, "Twitter tweet: GitHub Status (@githubstatus)"; got != want {
		t.Fatalf("tweet title = %q, want %q", got, want)
	}
	if got, want := first.Text, "First status update https://t.co/example"; got != want {
		t.Fatalf("tweet text = %q, want %q", got, want)
	}
	if first.UserID != d.UserID {
		t.Fatalf("tweet user ID = %d, want %d", first.UserID, d.UserID)
	}
	for key, want := range map[string]any{
		"type":      "tweet",
		"author":    "GitHub Status (@githubstatus)",
		"handle":    "@githubstatus",
		"published": "2026-08-11T20:06:56.000Z",
	} {
		if got := first.Metadata[key]; got != want {
			t.Errorf("tweet metadata %q = %#v, want %#v", key, got, want)
		}
	}
	if !strings.Contains(first.HTML, "<p>First status update") {
		t.Fatalf("tweet HTML is missing body text: %s", first.HTML)
	}
	if !strings.Contains(first.HTML, "https://pbs.twimg.com/media/example.jpg") {
		t.Fatalf("tweet HTML is missing media: %s", first.HTML)
	}
	if got, want := d.ExtraDocuments[1].URL, "https://x.com/githubstatus/status/456"; got != want {
		t.Fatalf("second tweet URL = %q, want %q", got, want)
	}
}

func TestExtractRenderedTweet(t *testing.T) {
	d := &document.Document{
		URL: "https://twitter.com/home",
		HTML: `<article data-testid="tweet">
			<div data-testid="User-Name">
				<a href="/alice"><span>Alice Example</span></a>
				<a href="/alice"><span>@alice</span></a>
			</div>
			<div data-testid="tweetText" lang="en" dir="auto">
				<span>Hello from the rendered feed </span><a href="/hashtag/golang">#golang</a>
			</div>
			<a href="/alice/status/987?s=20"><time datetime="2026-08-12T08:30:00.000Z">1h</time></a>
		</article>`,
	}

	state, err := (&TwitterExtractor{}).Extract(d)
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != types.ExtractorStop {
		t.Fatalf("Extract state = %v, want %v", state, types.ExtractorStop)
	}
	if len(d.ExtraDocuments) != 1 {
		t.Fatalf("ExtraDocuments length = %d, want 1", len(d.ExtraDocuments))
	}

	tweet := d.ExtraDocuments[0]
	if got, want := tweet.URL, "https://x.com/alice/status/987"; got != want {
		t.Fatalf("tweet URL = %q, want %q", got, want)
	}
	if got, want := tweet.Metadata["author"], "Alice Example (@alice)"; got != want {
		t.Fatalf("tweet author = %#v, want %#v", got, want)
	}
	if got, want := tweet.Metadata["published"], "2026-08-12T08:30:00.000Z"; got != want {
		t.Fatalf("tweet publication time = %#v, want %#v", got, want)
	}
	if !strings.Contains(tweet.Text, "Hello from the rendered feed") || !strings.Contains(tweet.Text, "#golang") {
		t.Fatalf("tweet text is incomplete: %q", tweet.Text)
	}
	if !strings.Contains(tweet.HTML, `href="https://twitter.com/hashtag/golang"`) {
		t.Fatalf("tweet HTML contains an unresolved link: %s", tweet.HTML)
	}
}

func TestExtractDirectTweetMetadataFallback(t *testing.T) {
	d := &document.Document{
		URL: "https://twitter.com/alice/status/654?s=20",
		HTML: `<html><head>
			<meta property="og:title" content="Alice Example (@alice) on X">
			<meta property="og:description" content="A tweet available without the rendered application.">
			<meta property="article:published_time" content="2026-08-12T09:45:00.000Z">
			<meta property="article:author" content="https://x.com/alice">
			<meta property="og:image" content="https://pbs.twimg.com/media/fallback.jpg">
		</head><body><div id="application"></div></body></html>`,
	}

	state, err := (&TwitterExtractor{}).Extract(d)
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != types.ExtractorStop {
		t.Fatalf("Extract state = %v, want %v", state, types.ExtractorStop)
	}
	if len(d.ExtraDocuments) != 1 {
		t.Fatalf("ExtraDocuments length = %d, want 1", len(d.ExtraDocuments))
	}
	tweet := d.ExtraDocuments[0]
	if got, want := tweet.URL, "https://x.com/alice/status/654"; got != want {
		t.Fatalf("tweet URL = %q, want %q", got, want)
	}
	if got, want := tweet.Text, "A tweet available without the rendered application."; got != want {
		t.Fatalf("tweet text = %q, want %q", got, want)
	}
	if got, want := tweet.Title, "Twitter tweet: Alice Example (@alice)"; got != want {
		t.Fatalf("tweet title = %q, want %q", got, want)
	}
	if got, want := tweet.Metadata["image"], "https://pbs.twimg.com/media/fallback.jpg"; got != want {
		t.Fatalf("tweet image = %#v, want %#v", got, want)
	}
	if !strings.Contains(tweet.HTML, "https://pbs.twimg.com/media/fallback.jpg") {
		t.Fatalf("tweet HTML is missing the fallback image: %s", tweet.HTML)
	}
}

func TestAuthorFromLegacyPageTitle(t *testing.T) {
	name, handle := authorFromPageTitle(`Alice Example on Twitter: "A tweet"`)
	if name != "Alice Example" || handle != "" {
		t.Fatalf("authorFromPageTitle returned (%q, %q), want (%q, %q)", name, handle, "Alice Example", "")
	}
}

func TestExtractSkipsFeedWhenNoTweetsFound(t *testing.T) {
	d := &document.Document{
		URL:  "https://x.com/home",
		HTML: `<html><body><div id="application"></div></body></html>`,
	}

	state, err := (&TwitterExtractor{}).Extract(d)
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != types.ExtractorStop {
		t.Fatalf("Extract state = %v, want %v", state, types.ExtractorStop)
	}
	if !d.SkipIndexing {
		t.Fatal("source feed was not marked to skip indexing")
	}
	if len(d.ExtraDocuments) != 0 {
		t.Fatalf("ExtraDocuments length = %d, want 0", len(d.ExtraDocuments))
	}
}

func TestExtractStopsForExtractedTweet(t *testing.T) {
	d := &document.Document{
		URL:      "https://x.com/alice/status/123",
		HTML:     `<article data-tweet-id="456"></article>`,
		Metadata: map[string]any{"type": "tweet"},
	}

	state, err := (&TwitterExtractor{}).Extract(d)
	if err != nil {
		t.Fatalf("Extract returned an error: %v", err)
	}
	if state != types.ExtractorStop {
		t.Fatalf("Extract state = %v, want %v", state, types.ExtractorStop)
	}
	if d.SkipIndexing {
		t.Fatal("extracted tweet was marked to skip indexing")
	}
	if len(d.ExtraDocuments) != 0 {
		t.Fatalf("ExtraDocuments length = %d, want 0", len(d.ExtraDocuments))
	}
}

func TestPreviewSanitizesTweetHTML(t *testing.T) {
	d := &document.Document{
		Title: `Tweet <script>alert("title")</script>`,
		HTML:  `<p onclick="alert(1)">Safe text<script>alert(2)</script><a href="javascript:alert(3)">bad link</a></p>`,
	}

	resp, state, err := (&TwitterExtractor{}).Preview(d)
	if err != nil {
		t.Fatalf("Preview returned an error: %v", err)
	}
	if state != types.ExtractorStop {
		t.Fatalf("Preview state = %v, want %v", state, types.ExtractorStop)
	}
	for _, disallowed := range []string{"<script", "onclick", "javascript:"} {
		if strings.Contains(strings.ToLower(resp.Content), disallowed) {
			t.Fatalf("preview contains %q: %s", disallowed, resp.Content)
		}
	}
	if !strings.Contains(resp.Content, "Safe text") {
		t.Fatalf("preview is missing tweet text: %s", resp.Content)
	}
}
