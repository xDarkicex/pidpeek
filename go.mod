module github.com/xDarkicex/pidpeek

go 1.25.7

require (
	github.com/ebitengine/purego v0.8.1
	github.com/xDarkicex/memory v0.0.0
	golang.org/x/sys v0.43.0
)

replace (
	github.com/xDarkicex/memory => ../memory
	github.com/xDarkicex/slabby => ../slabby
)
