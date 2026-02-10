/*
Wait for the new filesystem API planned by a Go proposal (from the core team).

This is currentyl just a stand-in for a real implementation.
That means: it's a quick hack!

It grows whenever I add some functionality needed for tests. Everything here
is quick and dirty right now.
*/
package locate

import (
	"os"
	"path/filepath"
)

// Return path for a resource file
func FileResource(item string, typ string) string {
	var path string
	switch typ {
	case "font":
		path = filepath.Join(os.Getenv("HOME"), "Library", "Fonts", item)
	}
	return path
}
