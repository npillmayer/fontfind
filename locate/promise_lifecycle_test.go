package locate_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/npillmayer/fontfind"
	"github.com/npillmayer/fontfind/fontregistry"
	"github.com/npillmayer/fontfind/locate"
)

func lifecyclePipeline(resolvers ...locate.FontLocatorWithContext) locate.ResolverPipeline {
	return locate.NewResolverPipeline(fontregistry.New(), resolvers...)
}

func lifecycleSuccess(context.Context, fontfind.Descriptor) (fontfind.ScalableFont, error) {
	return fontfind.FallbackFont(), nil
}

func TestPromiseRepeatedAwait(t *testing.T) {
	for _, outcome := range []string{"success", "miss", "canceled search"} {
		t.Run(outcome, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pipeline := lifecyclePipeline()
			if outcome == "success" {
				pipeline = lifecyclePipeline(lifecycleSuccess)
			} else if outcome == "canceled search" {
				cancel()
			}
			promise := pipeline.Resolve(ctx, fontfind.Descriptor{Pattern: "promise-probe"})
			first, firstErr := promise.Font()
			second, secondErr := promise.Font()
			if first.Name != second.Name || first.Path() != second.Path() || firstErr != secondErr {
				t.Fatalf("await changed result: first=(%q, %v), second=(%q, %v)", first.Name, firstErr, second.Name, secondErr)
			}
		})
	}
}

func TestPromiseConcurrentAwait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		pipeline := lifecyclePipeline(func(ctx context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
			<-release
			return lifecycleSuccess(ctx, d)
		})
		promise := pipeline.Resolve(context.Background(), fontfind.Descriptor{})
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				f, err := promise.Font()
				if f.Name != "Go-Regular.otf" || err != nil {
					t.Errorf("waiter received (%q, %v), want Go-Regular.otf without error", f.Name, err)
				}
			})
		}
		synctest.Wait()
		close(release)
		wg.Wait()
	})
}

// These tests intentionally let the promise go unconsumed. synctest reports a
// deadlock if the publisher remains blocked after the bubble's root returns.
func TestPromiseNoWaiter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		lifecyclePipeline(lifecycleSuccess).Resolve(context.Background(), fontfind.Descriptor{})
		synctest.Wait()
	})
}

func TestPromiseCanceledWaiter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		pipeline := lifecyclePipeline(func(ctx context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
			<-release
			return lifecycleSuccess(ctx, d)
		})
		promise := pipeline.Resolve(context.Background(), fontfind.Descriptor{})
		waitCtx, cancel := context.WithCancel(context.Background())
		cancel()
		f, err := promise.FontWithContext(waitCtx)
		if f.Name != "" || !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled wait returned (%q, %v)", f.Name, err)
		}
		close(release)
		synctest.Wait()
	})
}

func TestPromiseWaitCancellationDoesNotCancelSearch(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		pipeline := lifecyclePipeline(func(ctx context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
			<-release
			if err := ctx.Err(); err != nil {
				return fontfind.NullFont, err
			}
			return lifecycleSuccess(ctx, d)
		})
		promise := pipeline.Resolve(context.Background(), fontfind.Descriptor{})
		waitCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := promise.FontWithContext(waitCtx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("wait error = %v, want cancellation", err)
		}
		close(release)
		f, err := promise.Font()
		if f.Name != "Go-Regular.otf" || err != nil {
			t.Fatalf("later wait returned (%q, %v)", f.Name, err)
		}
	})
}

func TestPromiseSearchCancellationReachesResolver(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		started := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		pipeline := lifecyclePipeline(func(ctx context.Context, _ fontfind.Descriptor) (fontfind.ScalableFont, error) {
			close(started)
			<-ctx.Done()
			return fontfind.NullFont, ctx.Err()
		})
		promise := pipeline.Resolve(ctx, fontfind.Descriptor{})
		<-started
		cancel()
		f, err := promise.Font()
		if f.Name != "" || !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled search returned (%q, %v)", f.Name, err)
		}
	})
}

func TestPromiseCompletedResultWinsCancellation(t *testing.T) {
	promise := lifecyclePipeline(lifecycleSuccess).Resolve(nil, fontfind.Descriptor{})
	first, firstErr := promise.Font()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range 100 {
		f, err := promise.FontWithContext(ctx)
		if f.Name != first.Name || err != firstErr {
			t.Fatalf("completed result lost to cancellation: (%q, %v)", f.Name, err)
		}
	}
}

func TestPromiseNilWaitContext(t *testing.T) {
	f, err := lifecyclePipeline(lifecycleSuccess).Resolve(nil, fontfind.Descriptor{}).FontWithContext(nil)
	if f.Name != "Go-Regular.otf" || err != nil {
		t.Fatalf("nil-context wait returned (%q, %v)", f.Name, err)
	}
}

func TestPromiseWaitDeadlineLeavesOtherWaiterRunning(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		pipeline := lifecyclePipeline(func(ctx context.Context, d fontfind.Descriptor) (fontfind.ScalableFont, error) {
			<-release
			return lifecycleSuccess(ctx, d)
		})
		promise := pipeline.Resolve(context.Background(), fontfind.Descriptor{})
		otherDone := make(chan struct{})
		go func() {
			defer close(otherDone)
			f, err := promise.Font()
			if f.Name != "Go-Regular.otf" || err != nil {
				t.Errorf("independent waiter returned (%q, %v)", f.Name, err)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		f, err := promise.FontWithContext(ctx)
		if f.Name != "" || !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("deadline wait returned (%q, %v)", f.Name, err)
		}
		select {
		case <-otherDone:
			t.Error("other waiter returned before search completion")
		default:
		}
		close(release)
		<-otherDone
	})
}
