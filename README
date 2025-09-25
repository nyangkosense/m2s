m2s - generate S5 from markdown
===============================

`m2s` turns an arbitrary .md into a single-file S5 presentation.
Running the binary produces one self-contained HTML document with inline CSS
and JavaScript.

It's a lightweight solution to create portable Presentationslides, since it's a single .html file (S5), it runs in every Browser and is easily portable.

Build
-----

```
go mod tidy
go build -o m2s m2s.go
```

Run
---

```
./m2s slides.md            # writes slides.html next to slides.md
./m2s talk.md demo.html    # custom output file
```

Markdown is processed by `github.com/gomarkdown/markdown`, giving full
CommonMark coverage (tables, images, fenced code, etc.)

Input Format
--------------

```
---
title: My Talk
author: Speaker
---

# First Slide
Hello world

---

# Second Slide
- One point
- Another point
```

Front matter is optional; without it the first slide title becomes the deck title.
Each `---` line outside the front matter starts a new slide. Slides support
headings, paragraphs, bullet lists, fenced code blocks, and inline images via
`![alt](path-or-url)`.

Output Layout
-------------

The generated HTML already bundles the S5 UI (styles, navigation script,
print rules) so you can open it directly in a browser or host it as a static
file with no extra assets.
