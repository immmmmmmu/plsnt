package sitediff

import "errors"

// ErrDifferencesFound is returned by the CLI wrapper when --exit-code is
// set and the diff turned up differences. main() recognises this sentinel
// and exits with code 1 without printing it as an error — the diff body
// has already been printed on stdout.
var ErrDifferencesFound = errors.New("differences found")
