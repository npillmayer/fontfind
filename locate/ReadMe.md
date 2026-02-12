# locate

## Purpose

`locate` orchestrates font resolution across one or more providers.

It exposes async resolution primitives and a context-aware variant for cancellation/deadline control.

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
