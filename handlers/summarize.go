package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/EOEboh/mb-project-03-summarizer/ai"
	"github.com/yuin/goldmark"
)

// tmpl holds both named templates from index.html:
//
//	"index"  → full HTML page, served on GET /
//	"result" → HTML fragment, returned to HTMX on POST /summarize
//
// Both live in one file so they share CSS custom properties and stay in sync.
var tmpl = template.Must(template.ParseFiles("templates/index.html"))

// ── Data types ────────────────────────────────────────────────────────────────

// resultData is passed to the "result" template on every POST /summarize response.
// It covers both success and error states so the template stays simple.
type resultData struct {
	Error       string        // non-empty on failure; template shows error card
	Format      string        // "oneliner" | "bullets" | "full" | "all"
	OneLiner    template.HTML // safe HTML — produced by goldmark from trusted AI output
	Bullets     template.HTML
	FullSummary template.HTML
	WordCount   int // word count of the original input, displayed in the result card
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// Index serves the main summariser page.
func Index(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "index", nil); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// Summarize handles POST /summarize.
//
// ── What is new in Project 03 ─────────────────────────────────────────────────
//
// Projects 01 and 02 returned JSON to the browser. JavaScript processed it.
// This project returns an HTML fragment. HTMX swaps it directly into the DOM.
//
// The difference:
//
//	P01/P02:  Go → JSON string → JavaScript → DOM
//	P03:      Go → rendered HTML fragment → HTMX → DOM (no JS processing needed)
//
// When to return HTML fragments (HTMX pattern):
//   - The response maps cleanly to a UI component
//   - You want to avoid client-side rendering logic
//   - Server-side templates give you full control over markup
//
// When to return JSON:
//   - The same data feeds multiple UI components
//   - The client does post-processing (highlighting, formatting)
//   - You need the raw data for further client-side computation
//
// ─────────────────────────────────────────────────────────────────────────────
func Summarize(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.FormValue("text"))
	format := r.FormValue("format")
	if format == "" {
		format = "all"
	}

	// ── 1. Validate input ─────────────────────────────────────────────────
	if text == "" {
		renderFragment(w, resultData{Error: "Please paste some text to summarise."})
		return
	}

	wordCount := countWords(text)
	if wordCount < 20 {
		renderFragment(w, resultData{Error: fmt.Sprintf("Text is too short (%d words). Paste at least a few sentences.", wordCount)})
		return
	}
	if len(text) > 8000 {
		renderFragment(w, resultData{Error: "Text is too long. Please paste up to 8,000 characters."})
		return
	}

	// ── 2. Build messages ─────────────────────────────────────────────────
	messages := []ai.Message{
		{Role: "system", Content: buildPrompt(format)},
		{Role: "user", Content: text},
	}

	// ── 3. Call ai.Chat() ─────────────────────────────────────────────────
	// Same non-streaming pattern as Project 02.
	// ai.Chat() blocks until the full summary arrives, then we render it
	// into an HTML fragment and return it to HTMX.
	response, err := ai.Chat(ai.DefaultModel, messages)
	if err != nil {
		log.Printf("ai error: %v", err)
		renderFragment(w, resultData{Error: "AI request failed: " + err.Error()})
		return
	}

	// ── 4. Build result data and render fragment ──────────────────────────
	data := buildResultData(format, response, wordCount)
	renderFragment(w, data)
}

// ── Private helpers ───────────────────────────────────────────────────────────

// renderFragment writes the "result" named template as a Content-Type: text/html
// response. HTMX expects HTML back and swaps it into the target element.
//
// Note: Content-Type must be text/html here, not application/json.
// HTMX performs a standard HTML swap — sending JSON would render the raw JSON
// string as text content, not HTML.
func renderFragment(w http.ResponseWriter, data resultData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "result", data); err != nil {
		log.Printf("fragment render error: %v", err)
	}
}

// buildResultData converts the AI response string into typed resultData
// ready for template rendering.
func buildResultData(format, response string, wordCount int) resultData {
	switch format {
	case "oneliner":
		return resultData{Format: format, WordCount: wordCount, OneLiner: mdToHTML(response)}
	case "bullets":
		return resultData{Format: format, WordCount: wordCount, Bullets: mdToHTML(response)}
	case "full":
		return resultData{Format: format, WordCount: wordCount, FullSummary: mdToHTML(response)}
	default: // "all" — parse three sections from one response
		ol, bl, fs := parseSections(response)

		// Fallback: if the AI did not use the expected headings, treat the
		// entire response as a full summary rather than silently returning nothing.
		if ol == "" && bl == "" && fs == "" {
			fs = response
		}

		return resultData{
			Format:      "all",
			WordCount:   wordCount,
			OneLiner:    mdToHTML(ol),
			Bullets:     mdToHTML(bl),
			FullSummary: mdToHTML(fs),
		}
	}
}

// parseSections splits the "all" format response into its three sections.
//
// The AI is prompted to use ## ONE-LINER, ## KEY POINTS, ## FULL SUMMARY.
// We match these headings case-insensitively with partial matching to tolerate
// minor variations in the AI's output (e.g. "## ONE LINE SUMMARY").
func parseSections(response string) (oneliner, bullets, full string) {
	var (
		olBuf   strings.Builder
		blBuf   strings.Builder
		fsBuf   strings.Builder
		current string
	)

	for _, line := range strings.Split(response, "\n") {
		upper := strings.ToUpper(strings.TrimSpace(line))

		switch {
		case strings.Contains(upper, "ONE-LINER") || upper == "## ONE LINE":
			current = "oneliner"
		case strings.Contains(upper, "KEY POINTS") || strings.Contains(upper, "BULLET"):
			current = "bullets"
		case strings.Contains(upper, "FULL SUMMARY") || strings.Contains(upper, "COMPREHENSIVE"):
			current = "full"
		default:
			switch current {
			case "oneliner":
				olBuf.WriteString(line + "\n")
			case "bullets":
				blBuf.WriteString(line + "\n")
			case "full":
				fsBuf.WriteString(line + "\n")
			}
		}
	}

	return strings.TrimSpace(olBuf.String()),
		strings.TrimSpace(blBuf.String()),
		strings.TrimSpace(fsBuf.String())
}

// mdToHTML converts a markdown string to a template.HTML value using goldmark.
//
// Why goldmark instead of marked.js (which P02 used)?
// In P02 we returned JSON — the browser rendered the markdown client-side.
// In P03 we return an HTML fragment — Go must render the markdown server-side
// so HTMX receives ready-to-display HTML, not a raw markdown string.
//
// Why template.HTML?
// html/template escapes all string values by default to prevent XSS.
// template.HTML marks content as already-safe, bypassing escaping.
// We use it here because the content is produced by a local Ollama model
// (trusted source), not by end-user input.
func mdToHTML(md string) template.HTML {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.New().Convert([]byte(md), &buf); err != nil {
		// Fallback: escape and wrap in a paragraph
		return template.HTML("<p>" + template.HTMLEscapeString(md) + "</p>")
	}
	return template.HTML(buf.String())
}

// buildPrompt constructs the system prompt for the requested output format.
// The same five prompt engineering principles from Project 02 apply here,
// with structured output headings tuned for each format.
func buildPrompt(format string) string {
	const base = "You are an expert at analysing and distilling information clearly and concisely.\n\n"

	switch format {
	case "oneliner":
		return base + `Summarise the provided text in a single sentence.

Output: One sentence only. Maximum 30 words. No preamble such as "This text discusses" or "The article explains". Begin directly with the substance.`

	case "bullets":
		return base + `Summarise the provided text as a list of key points.

Output:
- Key point one
- Key point two
- Key point three
(5 to 7 bullet points. Each is one clear sentence. No nested bullets. No headers.)

Rules: Cover the most important ideas. Omit minor details.`

	case "full":
		return base + `Write a comprehensive summary of the provided text.

Output: Two or three paragraphs. 150 to 250 words. Cover the main argument, key supporting points, and conclusions. Match the tone of the original. No bullet points or headers.`

	default: // "all"
		return base + `Summarise the provided text in three formats. Use these exact section headings:

## ONE-LINER
A single sentence. Maximum 30 words. No preamble. Begin directly with the substance.

## KEY POINTS
- Key point one
- Key point two
- Key point three
(5 to 7 bullet points, one sentence each)

## FULL SUMMARY
Two or three paragraphs. 150 to 250 words. Main argument, key points, and conclusions.

Important: Use the exact headings shown above. Complete all three sections.`
	}
}

// countWords returns the number of whitespace-delimited words in s.
func countWords(s string) int {
	return len(strings.Fields(s))
}
