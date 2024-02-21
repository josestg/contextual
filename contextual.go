package contextual

import (
	"context"
)

// pair is a generic pair type.
type pair[F, S any] struct {
	f F
	s S
}

// pairOf returns a new pair.
func pairOf[F, S any](f F, s S) pair[F, S] {
	return pair[F, S]{f: f, s: s}
}

type textError string

func (e textError) Error() string { return string(e) }

// ErrChanClosed is an error returned when trying to receive from a closed channel.
const ErrChanClosed textError = "contextual: channel closed"

// Exec executes the function with the given context.
func Exec[T any](ctx context.Context, f func() (T, error)) (ret T, err error) {
	// Check if the context is already done, to prevent unnecessary work.
	select {
	case <-ctx.Done():
		return ret, ctx.Err()
	default:
	}

	res := make(chan pair[T, error], 1)
	go func() {
		res <- pairOf(f())
		close(res)
	}()

	select {
	case <-ctx.Done():
		return ret, ctx.Err()
	case r := <-res:
		return r.f, r.s
	}
}

// Recv receives a value from the channel with the given context.
func Recv[T any](ctx context.Context, ch <-chan T) (t T, err error) {
	select {
	case <-ctx.Done():
		return t, ctx.Err()
	case v, ok := <-ch:
		if !ok {
			return t, ErrChanClosed
		}
		return v, nil
	}
}

// Send sends a value to the channel with the given context.
func Send[T any](ctx context.Context, ch chan<- T, t T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- t:
		return nil
	}
}
