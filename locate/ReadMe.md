# locate

## Purpose

`locate` orchestrates font resolution across one or more providers.

It exposes async resolution primitives and a context-aware variant for cancellation/deadline control.

A promise retains one search result: repeated and concurrent awaits return the
same font and error. Completion does not require a waiter. `FontWithContext(ctx)`
cancels only that wait; other waiters and the search can continue. An observed
completed result takes precedence over wait cancellation, and a nil wait context
means `context.Background()`.

The context passed to `Resolve` or `ResolveFontLocWithContext` controls the search.
Resolvers must honor that context to interrupt work; adapting a legacy locator
does not make its I/O cancellable.

## API

- `type FontLocator`
- `type FontLocatorWithContext`
- `type FontPromise`
- `type FontRegistry`
- `type ResolverPipeline`
- `ResolveFontLoc(desc, resolvers...) FontPromise`
- `ResolveFontLocWithContext(ctx, desc, resolvers...) FontPromise`
- `NewResolverPipeline(reg, resolvers...) ResolverPipeline`
- `(ResolverPipeline).Resolve(ctx, desc) FontPromise`

Resolution flow:

1. Try registry cache.
2. Try resolvers in order.
3. Cache successful result.
4. Return fallback font with error when unresolved.

`ResolveFontLoc*` uses the global registry. Use `ResolverPipeline` when clients need their own registry instance.

## Example Applications

### 1. Resolve through a resolver chain

```go
promise := locate.ResolveFontLoc(desc, systemResolver, googleResolver, fallbackResolver)
sf, err := promise.Font()
```

For clients who need their own registry instance, create a custom pipeline:

```go
registry := fontregistry.New() // implements locate.FontRegistry
pipeline := locate.NewResolverPipeline(registry,
	systemResolver,
	googleResolver,
	fallbackResolver)
sf, err := pipeline.Resolve(context.Background(), desc).Font()
```

### 2. Context-aware waiting

```go
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

promise := locate.ResolveFontLocWithContext(ctx, desc, ctxResolver)
sf, err := promise.FontWithContext(ctx)
```

### 3. Custom registry pipeline

```go
reg := newClientRegistry() // implements locate.FontRegistry
pipeline := locate.NewResolverPipeline(reg, ctxResolver)
sf, err := pipeline.Resolve(context.Background(), desc).Font()
```
