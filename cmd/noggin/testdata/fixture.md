---
title: Fixture Document
tags: [noggin, fixture]
draft: true
---

# Reading Notes

A paragraph with *emphasis*, **strong emphasis** and `inline code`, long enough that it has to
wrap when the Document Viewer is only part of the terminal wide, which is the point of it
being here at all rather than being a short line that proves nothing about wrapping.

## Sections

### A table

| Field    | Meaning                  | Required |
|----------|--------------------------|----------|
| title    | Shown in the pane title  | no       |
| tags     | Filed under these        | no       |
| draft    | Hidden from listings     | no       |

### A code fence

```go
func Read(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return b, nil
}
```

### Nested lists

- Roots bound what is visible
  - The ad-hoc root comes from the launch argument
  - Starred roots persist across sessions
    - Stored as plain text, hand-editable
- Files are browsed within the active root
  1. Select
  2. Expand
  3. Activate

### Tasks

- [x] Render styled markdown
- [x] Hide frontmatter
- [ ] Find in document
- [ ] Help modal

### A blockquote

> Notes are not a database. They are a pile of paper you can find things in.
>
> — nobody in particular

---

## Filler

Everything below exists so the rendered document is taller than any pane it will be read in,
which is what makes scrolling observable.

1. The first thing
2. The second thing
3. The third thing
4. The fourth thing
5. The fifth thing
6. The sixth thing
7. The seventh thing
8. The eighth thing
9. The ninth thing
10. The tenth thing

Another wrapping paragraph, so that the tail of the document is not merely a list. It runs on
for a couple of lines to make sure the last screen has prose on it and can be told apart from
the first screen at a glance.

## The Last Section

The final line of the fixture document.
