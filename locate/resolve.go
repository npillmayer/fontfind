package locate

import (
	"context"
	"fmt"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/fontregistry"
)

// notFound returns an application error for a missing resource.
func notFound(res string) error {
	return fmt.Errorf("font not found: %v", res)
}

// fontPlusErr is a helper struct to exchange through channels.
type fontPlusErr struct {
	font fontfind.ScalableFont
	err  error
}

// FontPromise runs font searching asynchronously in the background.
// Font blocks until completion, and FontWithContext allows waiting with
// caller-controlled cancellation and deadlines.
type FontPromise interface {
	Font() (fontfind.ScalableFont, error)
	FontWithContext(ctx context.Context) (fontfind.ScalableFont, error)
}

type fontLoader struct {
	await func(ctx context.Context) (fontfind.ScalableFont, error)
}

func (loader fontLoader) Font() (fontfind.ScalableFont, error) {
	return loader.FontWithContext(context.Background())
}

func (loader fontLoader) FontWithContext(ctx context.Context) (fontfind.ScalableFont, error) {
	return loader.await(ctx)
}

// ResolveFontLoc resolves a scalable font using the given resolver chain.
//
// It first checks the global font registry cache. On a cache miss, resolvers are
// tried in the given order until one succeeds. A successful resolution is stored
// in the registry cache. If all resolvers fail, it returns the registry fallback
// font together with a not-found error.
//
// The search runs asynchronously and returns a FontPromise.
func ResolveFontLoc(desc fontfind.Descriptor, resolvers ...FontLocator) FontPromise {
	ctxResolvers := make([]FontLocatorWithContext, 0, len(resolvers))
	for _, r := range resolvers {
		ctxResolvers = append(ctxResolvers, adaptLocator(r))
	}
	return ResolveFontLocWithContext(context.Background(), desc, ctxResolvers...)
}

// ResolveFontLocWithContext is the context-aware variant of ResolveFontLoc.
// The search goroutine and resolver calls receive ctx.
func ResolveFontLocWithContext(ctx context.Context, desc fontfind.Descriptor, resolvers ...FontLocatorWithContext) FontPromise {
	if ctx == nil {
		ctx = context.Background()
	}
	ch := make(chan fontPlusErr)
	go func(ch chan<- fontPlusErr) {
		result := searchScalableFont(ctx, desc, resolvers)
		ch <- result
		close(ch)
	}(ch)
	loader := fontLoader{}
	// waitCtx is supplied by the caller when awaiting the promise.
	loader.await = func(waitCtx context.Context) (fontfind.ScalableFont, error) {
		select {
		case <-waitCtx.Done():
			return fontfind.NullFont, waitCtx.Err()
		case r := <-ch:
			return r.font, r.err
		}
	}
	return loader
}

func adaptLocator(r FontLocator) FontLocatorWithContext {
	return func(_ context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
		return r(d)
	}
}

func searchScalableFont(ctx context.Context, desc fontfind.Descriptor, resolvers []FontLocatorWithContext) (result fontPlusErr) {
	if err := ctx.Err(); err != nil {
		result.err = err
		return
	}
	name := fontregistry.NormalizeFontname(desc.Pattern, desc.Style, desc.Weight)
	if t, err := fontregistry.GlobalRegistry().GetFont(name); err == nil {
		result.font = t
		return
	}
	for _, resolver := range resolvers {
		if err := ctx.Err(); err != nil {
			result.err = err
			return
		}
		if f, err := resolver(ctx, desc); err == nil {
			fontregistry.GlobalRegistry().StoreFont(name, f)
			result.font = f
			return
		} else if ctxErr := ctx.Err(); ctxErr != nil {
			result.err = ctxErr
			return
		}
	}
	result.err = notFound(name)
	if f, err := fontregistry.GlobalRegistry().FallbackFont(); err == nil {
		result.font = f
	}
	return result
}
