package markdown

import (
	"strings"
	"testing"
)

func TestRenderHeading(t *testing.T) {
	html := Render("# Hello World")
	if !strings.Contains(html, "<h1") {
		t.Errorf("expected <h1> tag, got: %s", html)
	}
	if !strings.Contains(html, "Hello World") {
		t.Errorf("expected heading text, got: %s", html)
	}
	if !strings.Contains(html, `id="hello-world"`) {
		t.Errorf("expected slug id, got: %s", html)
	}
}

func TestRenderHeadingLevels(t *testing.T) {
	for i := 1; i <= 6; i++ {
		input := strings.Repeat("#", i) + " Heading"
		html := Render(input)
		tag := "<h" + string(rune('0'+i))
		if !strings.Contains(html, tag) {
			t.Errorf("expected %s in output, got: %s", tag, html)
		}
	}
}

func TestRenderParagraph(t *testing.T) {
	html := Render("This is a paragraph.")
	if !strings.Contains(html, "<p>") {
		t.Errorf("expected <p> tag, got: %s", html)
	}
	if !strings.Contains(html, "This is a paragraph.") {
		t.Errorf("expected paragraph text, got: %s", html)
	}
}

func TestRenderBold(t *testing.T) {
	html := Render("This is **bold** text.")
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("expected <strong> tag, got: %s", html)
	}
}

func TestRenderItalic(t *testing.T) {
	html := Render("This is *italic* text.")
	if !strings.Contains(html, "<em>italic</em>") {
		t.Errorf("expected <em> tag, got: %s", html)
	}
}

func TestRenderInlineCode(t *testing.T) {
	html := Render("Use `code` here.")
	if !strings.Contains(html, "<code>code</code>") {
		t.Errorf("expected <code> tag, got: %s", html)
	}
}

func TestRenderLink(t *testing.T) {
	html := Render("[Example](https://example.com)")
	if !strings.Contains(html, `<a href="https://example.com">Example</a>`) {
		t.Errorf("expected <a> tag, got: %s", html)
	}
}

func TestRenderImage(t *testing.T) {
	html := Render("![Alt text](https://example.com/img.png)")
	if !strings.Contains(html, `<img src="https://example.com/img.png" alt="Alt text">`) {
		t.Errorf("expected <img> tag, got: %s", html)
	}
}

func TestRenderUnorderedList(t *testing.T) {
	html := Render("- Item 1\n- Item 2\n- Item 3")
	if !strings.Contains(html, "<ul>") {
		t.Errorf("expected <ul> tag, got: %s", html)
	}
	if !strings.Contains(html, "<li>Item 1</li>") {
		t.Errorf("expected <li> tag, got: %s", html)
	}
	if !strings.Contains(html, "</ul>") {
		t.Errorf("expected closing </ul> tag, got: %s", html)
	}
}

func TestRenderOrderedList(t *testing.T) {
	html := Render("1. First\n2. Second\n3. Third")
	if !strings.Contains(html, "<ol>") {
		t.Errorf("expected <ol> tag, got: %s", html)
	}
	if !strings.Contains(html, "<li>First</li>") {
		t.Errorf("expected <li> tag, got: %s", html)
	}
}

func TestRenderCodeBlock(t *testing.T) {
	html := Render("```go\nfmt.Println(\"hello\")\n```")
	if !strings.Contains(html, "<pre><code") {
		t.Errorf("expected <pre><code> tag, got: %s", html)
	}
	if !strings.Contains(html, "fmt.Println") {
		t.Errorf("expected code content, got: %s", html)
	}
}

func TestRenderBlockquote(t *testing.T) {
	html := Render("> This is a quote")
	if !strings.Contains(html, "<blockquote>") {
		t.Errorf("expected <blockquote> tag, got: %s", html)
	}
	if !strings.Contains(html, "This is a quote") {
		t.Errorf("expected quote text, got: %s", html)
	}
}

func TestRenderHorizontalRule(t *testing.T) {
	html := Render("---")
	if !strings.Contains(html, "<hr>") {
		t.Errorf("expected <hr> tag, got: %s", html)
	}
}

func TestExtractText(t *testing.T) {
	md := "# Heading\n\nThis is **bold** and *italic* text with `code`."
	text := ExtractText(md)
	if strings.Contains(text, "#") {
		t.Errorf("heading marker should be removed, got: %s", text)
	}
	if strings.Contains(text, "**") {
		t.Errorf("bold markers should be removed, got: %s", text)
	}
	if strings.Contains(text, "*") {
		t.Errorf("italic markers should be removed, got: %s", text)
	}
	if strings.Contains(text, "`") {
		t.Errorf("code markers should be removed, got: %s", text)
	}
	if !strings.Contains(text, "Heading") {
		t.Errorf("heading text should be preserved, got: %s", text)
	}
	if !strings.Contains(text, "bold") {
		t.Errorf("bold text should be preserved, got: %s", text)
	}
}

func TestExtractTextLinks(t *testing.T) {
	md := "[Example](https://example.com) text"
	text := ExtractText(md)
	if !strings.Contains(text, "Example") {
		t.Errorf("link text should be preserved, got: %s", text)
	}
	if strings.Contains(text, "https://") {
		t.Errorf("URL should be removed, got: %s", text)
	}
}

func TestExtractTextImages(t *testing.T) {
	md := "![Alt text](https://example.com/img.png)"
	text := ExtractText(md)
	if !strings.Contains(text, "Alt text") {
		t.Errorf("alt text should be preserved, got: %s", text)
	}
	if strings.Contains(text, "https://") {
		t.Errorf("URL should be removed, got: %s", text)
	}
}

func TestTOC(t *testing.T) {
	md := `# Title

## Section One

### Subsection

## Section Two`
	headings := TOC(md)
	if len(headings) != 4 {
		t.Fatalf("expected 4 headings, got %d", len(headings))
	}
	if headings[0].Level != 1 || headings[0].Text != "Title" {
		t.Errorf("first heading wrong: %+v", headings[0])
	}
	if headings[1].Level != 2 || headings[1].Text != "Section One" {
		t.Errorf("second heading wrong: %+v", headings[1])
	}
	if headings[2].Level != 3 || headings[2].Text != "Subsection" {
		t.Errorf("third heading wrong: %+v", headings[2])
	}
}

func TestTOCWithCodeBlock(t *testing.T) {
	md := "# Real Heading\n\n```\n# Not a heading\n```\n\n## Another Heading"
	headings := TOC(md)
	if len(headings) != 2 {
		t.Fatalf("expected 2 headings (code block excluded), got %d", len(headings))
	}
	if headings[0].Text != "Real Heading" {
		t.Errorf("first heading wrong: %+v", headings[0])
	}
	if headings[1].Text != "Another Heading" {
		t.Errorf("second heading wrong: %+v", headings[1])
	}
}

func TestTOCSlug(t *testing.T) {
	headings := TOC("# Hello World!")
	if len(headings) != 1 {
		t.Fatalf("expected 1 heading, got %d", len(headings))
	}
	if headings[0].Slug != "hello-world" {
		t.Errorf("expected slug 'hello-world', got: %s", headings[0].Slug)
	}
}

func TestStats(t *testing.T) {
	md := `# Title

This is a paragraph with some words.

## Section

- Item 1
- Item 2

> A quote

` + "```" + `
code block
` + "```" + `

![image](img.png)

[link](url)`
	s := ComputeStats(md)
	if s.Headings != 2 {
		t.Errorf("expected 2 headings, got %d", s.Headings)
	}
	if s.Paragraphs < 2 {
		t.Errorf("expected at least 2 paragraphs, got %d", s.Paragraphs)
	}
	if s.ListItems != 2 {
		t.Errorf("expected 2 list items, got %d", s.ListItems)
	}
	if s.Blockquotes != 1 {
		t.Errorf("expected 1 blockquote, got %d", s.Blockquotes)
	}
	if s.CodeBlocks != 1 {
		t.Errorf("expected 1 code block, got %d", s.CodeBlocks)
	}
	if s.Images != 1 {
		t.Errorf("expected 1 image, got %d", s.Images)
	}
	if s.Links != 1 {
		t.Errorf("expected 1 link, got %d", s.Links)
	}
	if s.Words == 0 {
		t.Error("expected non-zero word count")
	}
	if s.ReadingTime < 1 {
		t.Error("expected at least 1 minute reading time")
	}
}

func TestStatsEmpty(t *testing.T) {
	s := ComputeStats("")
	if s.Words != 0 {
		t.Errorf("expected 0 words, got %d", s.Words)
	}
	if s.ReadingTime != 0 {
		t.Errorf("expected 0 reading time, got %d", s.ReadingTime)
	}
}

func TestHTMLToMarkdown(t *testing.T) {
	html := "<h1>Title</h1>\n<p>Some <strong>bold</strong> text.</p>"
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "# Title") {
		t.Errorf("expected heading, got: %s", md)
	}
	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected bold, got: %s", md)
	}
}

func TestHTMLToMarkdownLink(t *testing.T) {
	html := `<a href="https://example.com">Example</a>`
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "[Example](https://example.com)") {
		t.Errorf("expected link, got: %s", md)
	}
}

func TestHTMLToMarkdownImage(t *testing.T) {
	html := `<img src="https://example.com/img.png" alt="Alt">`
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "![Alt](https://example.com/img.png)") {
		t.Errorf("expected image, got: %s", md)
	}
}

func TestHTMLToMarkdownCodeBlock(t *testing.T) {
	html := "<pre><code>code here</code></pre>"
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "```") {
		t.Errorf("expected code fence, got: %s", md)
	}
	if !strings.Contains(md, "code here") {
		t.Errorf("expected code content, got: %s", md)
	}
}

func TestHTMLToMarkdownList(t *testing.T) {
	html := "<ul><li>Item 1</li><li>Item 2</li></ul>"
	md := HTMLToMarkdown(html)
	if !strings.Contains(md, "- Item 1") {
		t.Errorf("expected list item, got: %s", md)
	}
	if !strings.Contains(md, "- Item 2") {
		t.Errorf("expected list item, got: %s", md)
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Hello, World!", "hello-world"},
		{"Multiple   Spaces", "multiple-spaces"},
		{"UPPERCASE", "uppercase"},
		{"  Trim Me  ", "trim-me"},
		{"Special @#$% Chars", "special-chars"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRenderComplexDocument(t *testing.T) {
	md := `# Document Title

## Introduction

This is a **bold** paragraph with a [link](https://example.com).

### Subsection

- First item
- Second item with ` + "`code`" + `

> A blockquote

` + "```python" + `
print("hello")
` + "```"

	html := Render(md)
	// Should contain all elements
	checks := []string{
		"<h1", "Document Title",
		"<h2", "Introduction",
		"<h3", "Subsection",
		"<strong>bold</strong>",
		`<a href="https://example.com">link</a>`,
		"<ul>", "<li>First item</li>",
		"<blockquote>",
		`<pre><code class="language-python">`,
		`print(&quot;hello&quot;)`,
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("expected %q in output, got: %s", check, html)
		}
	}
}
