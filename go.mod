module github.com/kirill-scherba/teomon

replace github.com/kirill-scherba/teonet => ../teonet

replace github.com/kirill-scherba/trudp => ../trudp

replace github.com/kirill-scherba/teomon/teomon => ./teomon

go 1.16

require (
	github.com/kirill-scherba/teomon/teomon v0.0.0-00010101000000-000000000000 // indirect
	github.com/kirill-scherba/teonet v0.1.0
)
