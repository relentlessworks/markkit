package markdown

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Heading represents a single heading in a table of contents.
type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	Slug  string `json:"slug"`
}

// Stats holds document statistics.
type Stats struct {
	Words       int `json:"words"`
	Chars       int `json:"chars"`
	Runes       int `json:"runes"`
	Lines       int `json:"lines"`
	Paragraphs  int `json:"paragraphs"`
	Headings    int `json:"headings"`
	CodeBlocks  int `json:"code_blocks"`
	Links       int `json:"links"`
	Images      int `json:"images"`
	ListItems   int `json:"list_items"`
	Blockquotes int `json:"blockquotes"`
	ReadingTime int `json:"reading_time_min"`
}

var (
	// Heading regex: matches # Heading text
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	// Code fence regex
	codeFenceRe = regexp.MustCompile("^```")
	// Link regex: [text](url)
	linkRe = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
	// Image regex: ![alt](url)
	imageRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	// List item regex
	listItemRe = regexp.MustCompile(`^[\s]*[-*+]\s+|^[\s]*\d+\.\s+`)
	// Blockquote regex
	blockquoteRe = regexp.MustCompile(`^>\s+`)
	// HTML tag regex for stripping
	htmlTagRe = regexp.MustCompile(`<[^>]+>`)
	// Multiple whitespace
	multiSpaceRe = regexp.MustCompile(`[ \t]+`)
	// Multiple newlines
	multiNewlineRe = regexp.MustCompile(`\n{3,}`)
)

// Render converts Markdown to HTML.
func Render(md string) string {
	lines := strings.Split(md, "\n")
	var out strings.Builder
	inCodeBlock := false
	inList := false
	listType := "" // "ul" or "ol"

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code fence handling
		if codeFenceRe.MatchString(trimmed) {
			if inCodeBlock {
				out.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				lang := strings.TrimPrefix(trimmed, "```")
				lang = strings.TrimSpace(lang)
				if lang != "" {
					out.WriteString(fmt.Sprintf("<pre><code class=\"language-%s\">\n", lang))
				} else {
					out.WriteString("<pre><code>\n")
				}
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			out.WriteString(escapeHTML(line))
			out.WriteString("\n")
			continue
		}

		// Close list if we hit a non-list line
		if inList && !listItemRe.MatchString(line) && trimmed != "" {
			if listType == "ul" {
				out.WriteString("</ul>\n")
			} else {
				out.WriteString("</ol>\n")
			}
			inList = false
		}

		// Empty line
		if trimmed == "" {
			continue
		}

		// Headings
		if m := headingRe.FindStringSubmatch(trimmed); m != nil {
			level := len(m[1])
			text := inlineFormat(m[2])
			slug := slugify(m[2])
			out.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", level, slug, text, level))
			continue
		}

		// Blockquote
		if blockquoteRe.MatchString(trimmed) {
			text := inlineFormat(strings.TrimPrefix(trimmed, "> "))
			out.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", text))
			continue
		}

		// Unordered list item
		if m := regexp.MustCompile(`^[-*+]\s+(.+)`).FindStringSubmatch(trimmed); m != nil {
			if !inList || listType != "ul" {
				if inList {
					out.WriteString("</ol>\n")
				}
				out.WriteString("<ul>\n")
				inList = true
				listType = "ul"
			}
			out.WriteString(fmt.Sprintf("<li>%s</li>\n", inlineFormat(m[1])))
			continue
		}

		// Ordered list item
		if m := regexp.MustCompile(`^\d+\.\s+(.+)`).FindStringSubmatch(trimmed); m != nil {
			if !inList || listType != "ol" {
				if inList {
					out.WriteString("</ul>\n")
				}
				out.WriteString("<ol>\n")
				inList = true
				listType = "ol"
			}
			out.WriteString(fmt.Sprintf("<li>%s</li>\n", inlineFormat(m[1])))
			continue
		}

		// Horizontal rule
		if regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`).MatchString(trimmed) {
			out.WriteString("<hr>\n")
			continue
		}

		// Paragraph (default)
		out.WriteString(fmt.Sprintf("<p>%s</p>\n", inlineFormat(trimmed)))
	}

	// Close any open list
	if inList {
		if listType == "ul" {
			out.WriteString("</ul>\n")
		} else {
			out.WriteString("</ol>\n")
		}
	}

	// Close any open code block
	if inCodeBlock {
		out.WriteString("</code></pre>\n")
	}

	return out.String()
}

// ExtractText strips all Markdown formatting and returns plain text.
func ExtractText(md string) string {
	lines := strings.Split(md, "\n")
	var out strings.Builder
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if codeFenceRe.MatchString(trimmed) {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		// Remove heading markers
		if m := headingRe.FindStringSubmatch(trimmed); m != nil {
			out.WriteString(m[2])
			out.WriteString("\n")
			continue
		}

		// Remove list markers
		line = listItemRe.ReplaceAllString(line, "")
		// Remove blockquote markers
		line = blockquoteRe.ReplaceAllString(line, "")
		// Remove horizontal rules
		if regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`).MatchString(trimmed) {
			continue
		}

		// Remove image syntax, keep alt text
		line = imageRe.ReplaceAllString(line, "$1")
		// Remove link syntax, keep text
		line = linkRe.ReplaceAllString(line, "$1")
		// Remove emphasis markers
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.ReplaceAll(line, "*", "")
		line = strings.ReplaceAll(line, "_", "")
		line = strings.ReplaceAll(line, "`", "")

		out.WriteString(strings.TrimSpace(line))
		out.WriteString("\n")
	}

	// Collapse multiple blank lines
	result := out.String()
	result = multiNewlineRe.ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}

// TOC generates a table of contents from Markdown headings.
func TOC(md string) []Heading {
	lines := strings.Split(md, "\n")
	var headings []Heading
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if codeFenceRe.MatchString(trimmed) {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		if m := headingRe.FindStringSubmatch(trimmed); m != nil {
			level := len(m[1])
			text := strings.TrimSpace(m[2])
			slug := slugify(text)
			headings = append(headings, Heading{
				Level: level,
				Text:  text,
				Slug:  slug,
			})
		}
	}

	return headings
}

// ComputeStats returns statistics about the Markdown document.
func ComputeStats(md string) Stats {
	lines := strings.Split(md, "\n")
	stats := Stats{}
	inCodeBlock := false
	blankLine := true

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		stats.Lines++

		if codeFenceRe.MatchString(trimmed) {
			inCodeBlock = !inCodeBlock
			if !inCodeBlock {
				stats.CodeBlocks++
			}
			continue
		}

		if inCodeBlock {
			continue
		}

		// Count headings
		if headingRe.MatchString(trimmed) {
			stats.Headings++
		}

		// Count images
		stats.Images += len(imageRe.FindAllString(trimmed, -1))

		// Count links (subtract images since images match link pattern too)
		linkCount := len(linkRe.FindAllString(trimmed, -1))
		imgCount := len(imageRe.FindAllString(trimmed, -1))
		stats.Links += linkCount - imgCount

		// Count list items
		if listItemRe.MatchString(line) {
			stats.ListItems++
		}

		// Count blockquotes
		if blockquoteRe.MatchString(trimmed) {
			stats.Blockquotes++
		}

		// Count paragraphs (transition from blank to non-blank)
		if trimmed == "" {
			blankLine = true
		} else {
			if blankLine {
				stats.Paragraphs++
			}
			blankLine = false
		}
	}

	// Word and character count from extracted text
	plain := ExtractText(md)
	words := strings.Fields(plain)
	stats.Words = len(words)
	stats.Chars = len(plain)
	stats.Runes = utf8.RuneCountInString(plain)

	// Reading time: ~200 words per minute, minimum 1
	if stats.Words > 0 {
		stats.ReadingTime = (stats.Words + 199) / 200
	}

	return stats
}

// HTMLToMarkdown converts simple HTML to Markdown.
func HTMLToMarkdown(html string) string {
	// Remove HTML comments
	html = regexp.MustCompile(`<!--[\s\S]*?-->`).ReplaceAllString(html, "")

	// Convert headings
	for i := 6; i >= 1; i-- {
		prefix := strings.Repeat("#", i)
		re := regexp.MustCompile(fmt.Sprintf(`(?i)<h%d[^>]*>(.*?)</h%d>`, i, i))
		html = re.ReplaceAllStringFunc(html, func(s string) string {
			m := re.FindStringSubmatch(s)
			if m != nil {
				return "\n" + prefix + " " + stripTags(m[1]) + "\n"
			}
			return s
		})
	}

	// Convert links
	linkRe := regexp.MustCompile(`(?i)<a\s+[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	html = linkRe.ReplaceAllString(html, "[$2]($1)")

	// Convert images
	imgRe := regexp.MustCompile(`(?i)<img\s+[^>]*alt="([^"]*)"[^>]*src="([^"]*)"[^>]*/?>`)
	html = imgRe.ReplaceAllString(html, "![$1]($2)")
	imgRe2 := regexp.MustCompile(`(?i)<img\s+[^>]*src="([^"]*)"[^>]*alt="([^"]*)"[^>]*/?>`)
	html = imgRe2.ReplaceAllString(html, "![$2]($1)")

	// Convert bold
	html = regexp.MustCompile(`(?i)<strong>(.*?)</strong>`).ReplaceAllString(html, "**$1**")
	html = regexp.MustCompile(`(?i)<b>(.*?)</b>`).ReplaceAllString(html, "**$1**")

	// Convert italic
	html = regexp.MustCompile(`(?i)<em>(.*?)</em>`).ReplaceAllString(html, "*$1*")
	html = regexp.MustCompile(`(?i)<i>(.*?)</i>`).ReplaceAllString(html, "*$1*")

	// Convert pre/code blocks (must come before inline code)
	html = regexp.MustCompile(`(?i)<pre><code[^>]*>([\s\S]*?)</code></pre>`).ReplaceAllStringFunc(html, func(s string) string {
		m := regexp.MustCompile(`(?i)<pre><code[^>]*>([\s\S]*?)</code></pre>`).FindStringSubmatch(s)
		if m != nil {
			return "```\n" + unescapeHTML(m[1]) + "\n```"
		}
		return s
	})

	// Convert inline code
	html = regexp.MustCompile(`(?i)<code>(.*?)</code>`).ReplaceAllString(html, "`$1`")

	// Convert blockquotes
	html = regexp.MustCompile(`(?i)<blockquote>([\s\S]*?)</blockquote>`).ReplaceAllStringFunc(html, func(s string) string {
		m := regexp.MustCompile(`(?i)<blockquote>([\s\S]*?)</blockquote>`).FindStringSubmatch(s)
		if m != nil {
			inner := strings.TrimSpace(stripTags(m[1]))
			lines := strings.Split(inner, "\n")
			for i, l := range lines {
				lines[i] = "> " + l
			}
			return strings.Join(lines, "\n")
		}
		return s
	})

	// Convert list items
	html = regexp.MustCompile(`(?i)<li>(.*?)</li>`).ReplaceAllStringFunc(html, func(s string) string {
		m := regexp.MustCompile(`(?i)<li>(.*?)</li>`).FindStringSubmatch(s)
		if m != nil {
			return "- " + strings.TrimSpace(stripTags(m[1])) + "\n"
		}
		return s
	})

	// Remove list wrappers
	html = regexp.MustCompile(`(?i)</?[uo]l[^>]*>`).ReplaceAllString(html, "")

	// Convert paragraphs
	html = regexp.MustCompile(`(?i)<p[^>]*>(.*?)</p>`).ReplaceAllStringFunc(html, func(s string) string {
		m := regexp.MustCompile(`(?i)<p[^>]*>(.*?)</p>`).FindStringSubmatch(s)
		if m != nil {
			return strings.TrimSpace(stripTags(m[1])) + "\n\n"
		}
		return s
	})

	// Convert <br> to newlines
	html = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(html, "\n")

	// Convert <hr> to ---
	html = regexp.MustCompile(`(?i)<hr\s*/?>`).ReplaceAllString(html, "\n---\n")

	// Strip remaining tags
	html = stripTags(html)

	// Unescape HTML entities
	html = unescapeHTML(html)

	// Clean up whitespace
	html = multiSpaceRe.ReplaceAllString(html, " ")
	html = multiNewlineRe.ReplaceAllString(html, "\n\n")

	return strings.TrimSpace(html)
}

// inlineFormat processes inline Markdown formatting (bold, italic, code, links, images).
func inlineFormat(s string) string {
	// Images: ![alt](url)
	s = imageRe.ReplaceAllString(s, `<img src="$2" alt="$1">`)
	// Links: [text](url)
	s = linkRe.ReplaceAllString(s, `<a href="$2">$1</a>`)
	// Bold: **text** or __text__
	s = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(s, "<strong>$1</strong>")
	s = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(s, "<strong>$1</strong>")
	// Italic: *text* or _text_
	s = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(s, "<em>$1</em>")
	s = regexp.MustCompile(`\b_([^_]+)_\b`).ReplaceAllString(s, "<em>$1</em>")
	// Inline code: `code`
	s = regexp.MustCompile("`([^`]+)`").ReplaceAllString(s, "<code>$1</code>")
	return s
}

// slugify creates a URL-safe slug from heading text.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric except hyphens
	s = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(s, "")
	// Collapse multiple hyphens
	s = regexp.MustCompile(`-{2,}`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// escapeHTML escapes HTML special characters.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// unescapeHTML reverses HTML entity escaping.
func unescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&#x27;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	return s
}

// stripTags removes all HTML tags from a string.
func stripTags(s string) string {
	return htmlTagRe.ReplaceAllString(s, "")
}
