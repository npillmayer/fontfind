/*
Package locate defines resolver function types and async font resolution helpers.
*/
package locate

import (
	"context"

	"github.com/npillmayer/fontfind"
)

// FontLocator resolves a scalable font for a descriptor.
type FontLocator func(fontfind.Descriptor) (fontfind.ScalableFont, error)

// FontLocatorWithContext is a context-aware variant of FontLocator.
// Implementations should respect cancellation/deadlines of ctx if possible.
type FontLocatorWithContext func(context.Context, fontfind.Descriptor) (fontfind.ScalableFont, error)
