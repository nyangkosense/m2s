package main

import (
	_ "embed"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

//go:embed slides.css.tmpl
var slidesCSS string

//go:embed outline.css.tmpl
var outlineCSS string

//go:embed print.css.tmpl
var printCSS string

//go:embed slides.js.tmpl
var slidesJS string

//go:embed page.tmpl
var pageTmpl string

type Slide struct {
	T string
	C string
}

type Deck struct {
	T string
	A string
	S []Slide
}

var frontMatter = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n?`)
var sectionSplit = regexp.MustCompile(`(?m)^\s*---\s*$`)
var nonWord = regexp.MustCompile(`[^a-z0-9]+`)

var mdExtensions = parser.CommonExtensions |
	parser.AutoHeadingIDs |
	parser.NoEmptyLineBeforeBlock

var mdRenderer = mdhtml.NewRenderer(mdhtml.RendererOptions{
	Flags: mdhtml.CommonFlags | mdhtml.HrefTargetBlank,
})

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func read(f string) string {
	b, err := os.ReadFile(f)
	die(err)
	return string(b)
}

func write(f, s string) {
	d := filepath.Dir(f)
	if d != "." {
		die(os.MkdirAll(d, 0o755))
	}
	err := os.WriteFile(f, []byte(s), 0o644)
	die(err)
}

func parse(s string) Deck {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)

	deck := Deck{}
	body := s
	if m := frontMatter.FindStringSubmatch(s); len(m) == 2 {
		deck.T, deck.A = meta(m[1])
		body = strings.TrimSpace(s[len(m[0]):])
	}

	chunks := split(body)
	for _, chunk := range chunks {
		if chunk == "" {
			continue
		}
		lines := strings.Split(chunk, "\n")
		head := ""
		rest := []string{}
		if len(lines) > 0 {
			head = strings.TrimSpace(lines[0])
			rest = lines[1:]
		}

		title := head
		contentLines := rest
		if strings.HasPrefix(head, "#") {
			title = strings.TrimSpace(strings.TrimLeft(head, "# "))
		} else if title != "" {
			contentLines = lines
		}

		bodyHTML := md(strings.Join(contentLines, "\n"))
		if strings.TrimSpace(bodyHTML) == "" {
			bodyHTML = md(title)
		}

		deck.S = append(deck.S, Slide{T: title, C: bodyHTML})
	}

	if deck.T == "" && len(deck.S) > 0 {
		deck.T = deck.S[0].T
	}

	return deck
}

func split(s string) []string {
	raw := sectionSplit.Split(s, -1)
	slides := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			slides = append(slides, trimmed)
		}
	}
	return slides
}

func meta(s string) (string, string) {
	lines := strings.Split(s, "\n")
	title := ""
	author := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) >= 6 && strings.EqualFold(line[:6], "title:") {
			title = strings.TrimSpace(line[6:])
		}
		if len(line) >= 7 && strings.EqualFold(line[:7], "author:") {
			author = strings.TrimSpace(line[7:])
		}
	}
	return title, author
}

func md(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	p := parser.NewWithExtensions(mdExtensions)
	return string(markdown.ToHTML([]byte(trimmed), p, mdRenderer))
}

func slug(s string) string {
	low := strings.ToLower(s)
	low = nonWord.ReplaceAllString(low, "-")
	return strings.Trim(low, "-")
}

func render(d Deck) string {
	data := struct {
		Title      string
		Author     string
		ScreenCSS  template.CSS
		OutlineCSS template.CSS
		PrintCSS   template.CSS
		JS         template.JS
		Slides     []struct {
			Title   string
			Body    template.HTML
			ID      string
			Current bool
		}
	}{
		Title:      d.T,
		Author:     d.A,
		ScreenCSS:  template.CSS(slidesCSS),
		OutlineCSS: template.CSS(outlineCSS),
		PrintCSS:   template.CSS(printCSS),
		JS:         template.JS(slidesJS),
	}

	for i, s := range d.S {
		id := slug(s.T)
		if id == "" {
			id = "slide-" + strconv.Itoa(i+1)
		}
		data.Slides = append(data.Slides, struct {
			Title   string
			Body    template.HTML
			ID      string
			Current bool
		}{
			Title:   s.T,
			Body:    template.HTML(s.C),
			ID:      id,
			Current: i == 0,
		})
	}

	t := template.Must(template.New("page").Parse(pageTmpl))
	var b strings.Builder
	die(t.Execute(&b, data))
	return b.String()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: m2s input.md [output.html]")
	}

	in := os.Args[1]
	base := strings.TrimSuffix(in, filepath.Ext(in))
	if base == "" {
		base = in
	}
	out := base + ".html"
	if len(os.Args) > 2 {
		out = os.Args[2]
	}

	raw := read(in)
	deck := parse(raw)
	html := render(deck)
	write(out, html)
}
