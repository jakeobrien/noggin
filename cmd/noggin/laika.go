package main

import _ "embed"

// laika is the ASCII art the Document Viewer shows when no document is open.
//
//go:embed laika.txt
var laika string
