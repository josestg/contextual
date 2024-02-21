package contextual

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTextError_Error(t *testing.T) {
	e := textError("some error")
	if e.Error() != "some error" {
		t.Errorf("expect 'some error'; got %s", e.Error())
	}
}

func TestExec(t *testing.T) {
	dummyErr := errors.New("some error")

	timed := func(delay time.Duration) func() (int, error) {
		return func() (int, error) {
			time.Sleep(delay)
			return 42, dummyErr
		}
	}

	t.Run("no cancel", func(t *testing.T) {
		ctx := context.Background()
		got, err := Exec(ctx, timed(time.Millisecond))
		if !errors.Is(err, dummyErr) {
			t.Errorf("expect %v; got %v", dummyErr, err)
		}
		if got != 42 {
			t.Errorf("expect 42; got %d", got)
		}
	})

	t.Run("deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		got, err := Exec(ctx, timed(time.Second))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expect %v; got %v", context.DeadlineExceeded, err)
		}
		if got != 0 {
			t.Errorf("expect 0; got %d", got)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(100*time.Millisecond, cancel)
		got, err := Exec(ctx, timed(time.Second))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expect %v; got %v", context.DeadlineExceeded, err)
		}
		if got != 0 {
			t.Errorf("expect 0; got %d", got)
		}
	})

	t.Run("cancel before start", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel before starting the function.

		got, err := Exec(ctx, timed(time.Second))
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expect %v; got %v", context.DeadlineExceeded, err)
		}
		if got != 0 {
			t.Errorf("expect 0; got %d", got)
		}
	})
}

func TestRecv(t *testing.T) {
	t.Run("no cancel", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 42
		close(ch)

		ctx := context.Background()
		got, err := Recv(ctx, ch)
		if err != nil {
			t.Errorf("expect nil; got %v", err)
		}
		if got != 42 {
			t.Errorf("expect 42; got %d", got)
		}
	})

	t.Run("channel closed", func(t *testing.T) {
		ch := make(chan int)
		close(ch)

		ctx := context.Background()
		got, err := Recv(ctx, ch)
		if !errors.Is(err, ErrChanClosed) {
			t.Errorf("expect %v; got %v", ErrChanClosed, err)
		}
		if got != 0 {
			t.Errorf("expect 0; got %d", got)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ch := make(chan int)
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(100*time.Millisecond, cancel)

		got, err := Recv(ctx, ch)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expect %v; got %v", context.Canceled, err)
		}
		if got != 0 {
			t.Errorf("expect 0; got %d", got)
		}
	})

	t.Run("cancel before start", func(t *testing.T) {
		ch := make(chan int)
		cancel := func() {}
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel before starting the function.

		got, err := Recv(ctx, ch)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expect %v; got %v", context.Canceled, err)
		}
		if got != 0 {
			t.Errorf("expect 0; got %d", got)
		}
	})
}

func TestSend(t *testing.T) {
	t.Run("no cancel", func(t *testing.T) {
		ch := make(chan int, 1)
		ctx := context.Background()
		err := Send(ctx, ch, 42)
		if err != nil {
			t.Errorf("expect nil; got %v", err)
		}
		if got := <-ch; got != 42 {
			t.Errorf("expect 42; got %d", got)
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ch := make(chan int)
		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(100*time.Millisecond, cancel)

		err := Send(ctx, ch, 42)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expect %v; got %v", context.Canceled, err)
		}
	})

	t.Run("cancel before start", func(t *testing.T) {
		ch := make(chan int)
		cancel := func() {}
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel before starting the function.

		err := Send(ctx, ch, 42)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expect %v; got %v", context.Canceled, err)
		}
	})
}
