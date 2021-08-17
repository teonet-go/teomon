module github.com/kirill-scherba/teomon

// replace github.com/kirill-scherba/teonet => ../teonet

// replace github.com/kirill-scherba/trudp => ../trudp

replace github.com/kirill-scherba/teomon/teomon => ./teomon

go 1.16

require (
	github.com/kirill-scherba/teomon/teomon v0.0.0-20210628141718-30d2c073283e
	github.com/kirill-scherba/teonet v0.2.22
	github.com/kirill-scherba/trudp v0.1.1 // indirect
)
