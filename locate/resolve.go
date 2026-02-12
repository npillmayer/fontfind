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

// FontRegistry is the cache contract required by ResolverPipeline.
type FontRegistry interface {
	GetFont(string) (fontfind.ScalableFont, error)
	StoreFont(string, fontfind.ScalableFont)
	FallbackFont() (fontfind.ScalableFont, error)
}

// ResolverPipeline orchestrates resolver execution with a configurable registry.
type ResolverPipeline struct {
	registry  FontRegistry
	resolvers []FontLocatorWithContext
}

// NewResolverPipeline constructs a resolver driver with an optional custom registry.
// If reg is nil, the global registry singleton is used.
func NewResolverPipeline(reg FontRegistry, resolvers ...FontLocatorWithContext) ResolverPipeline {
	if reg == nil {
		reg = fontregistry.GlobalRegistry()
	}
	rs := make([]FontLocatorWithContext, len(resolvers))
	copy(rs, resolvers)
	return ResolverPipeline{
		registry:  reg,
		resolvers: rs,
	}
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
	return NewResolverPipeline(nil, ctxResolvers...).Resolve(context.Background(), desc)
}

// ResolveFontLocWithContext is the context-aware variant of ResolveFontLoc.
// The search goroutine and resolver calls receive ctx.
func ResolveFontLocWithContext(ctx context.Context, desc fontfind.Descriptor, resolvers ...FontLocatorWithContext) FontPromise {
	return NewResolverPipeline(nil, resolvers...).Resolve(ctx, desc)
}

// Resolve resolves a font request asynchronously using this pipeline's registry and resolvers.
func (pipeline ResolverPipeline) Resolve(ctx context.Context, desc fontfind.Descriptor) FontPromise {
	if ctx == nil {
		ctx = context.Background()
	}
	registry := pipeline.registry
	if registry == nil {
		registry = fontregistry.GlobalRegistry()
	}
	ch := make(chan fontPlusErr)
	go func(ch chan<- fontPlusErr) {
		result := searchScalableFont(ctx, registry, desc, pipeline.resolvers)
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

func searchScalableFont(ctx context.Context, registry FontRegistry, desc fontfind.Descriptor, resolvers []FontLocatorWithContext) (result fontPlusErr) {
	if err := ctx.Err(); err != nil {
		result.err = err
		return
	}
	if registry == nil {
		registry = fontregistry.GlobalRegistry()
	}
	name := fontregistry.NormalizeFontname(desc.Pattern, desc.Style, desc.Weight)
	if t, err := registry.GetFont(name); err == nil {
		result.font = t
		return
	}
	for _, resolver := range resolvers {
		if err := ctx.Err(); err != nil {
			result.err = err
			return
		}
		if f, err := resolver(ctx, desc); err == nil {
			registry.StoreFont(name, f)
			result.font = f
			return
		} else if ctxErr := ctx.Err(); ctxErr != nil {
			result.err = ctxErr
			return
		}
	}
	result.err = notFound(name)
	if f, err := registry.FallbackFont(); err == nil {
		result.font = f
	}
	return result
}
