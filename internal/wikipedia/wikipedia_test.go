package wikipedia

// sampleTitles are the S5 judging sample, in spec order (mirrors
// cmd/wikispike's fetch_test.go copy — the two packages test the pipeline
// against the same fixed article set independently, without either
// importing the other's test-only data).
var sampleTitles = []string{
	"Why is there anything at all?",
	"E (mathematical constant)",
	"List of programming languages by type",
	"Major Arcana",
}
