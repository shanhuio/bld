package dock

import "errors"

// ErrFileNotFound is used when a file is not found in a container.
var ErrFileNotFound = errors.New("file not found")
