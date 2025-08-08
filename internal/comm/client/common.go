package client

import (
	"os"
	"path/filepath"
)

var socket = filepath.Join(os.TempDir(), "elephant.sock")

const done = 255
