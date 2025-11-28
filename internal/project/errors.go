package project

import "errors"

// ErrNotTracksProject indicates the directory is not a Tracks project.
var ErrNotTracksProject = errors.New("not a tracks project")
