# locate

## Purpose

`locate` orchestrates font resolution across one or more providers.

It exposes async resolution primitives and a context-aware variant for cancellation/deadline control.

## API

- `type FontLocator`
- `type FontLocatorWithContext`
- `type FontPromise`
- `ResolveFontLoc(desc, resolvers...) FontPromise`
- `ResolveFontLocWithContext(ctx, desc, resolvers...) FontPromise`

Resolution flow:

1. Try registry cache.
2. Try resolvers in order.
3. Cache successful result.
4. Return fallback font with error when unresolved.

## Example Applications

### 1. Resolve through a resolver chain

```go
promise := locate.ResolveFontLoc(desc, systemResolver, googleResolver, fallbackResolver)
sf, err := promise.Font()
```

### 2. Context-aware waiting

```go
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()

promise := locate.ResolveFontLocWithContext(ctx, desc, ctxResolver)
sf, err := promise.FontWithContext(ctx)
```
