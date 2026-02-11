/*
Wait for the new filesystem API planned by a Go proposal (from the core team).

This is currentyl just a stand-in for a real implementation.
That means: it's a quick hack!

It grows whenever I add some functionality needed for tests. Everything here
is quick and dirty right now.
*/
package locate

import (
	"github.com/npillmayer/fontfind"
)

type FontLocator func(fontfind.Descriptor) (fontfind.ScalableFont, error)
