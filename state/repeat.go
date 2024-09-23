package state

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Quantifier is a Count state quantifier.
type Quantifier struct {
	min uint
	max uint
}

// CountBetween returns a new Quantifier with min and max runes to match.
// Cases:
//   - min == max: match exact count of runes
//   - min < max: match any count of runes in range [min, max]
//   - min == 0: match any count of runes in range [0, max]
//   - max == math.MaxUint: match any count of runes in range [min, infinity(EOF)]
func CountBetween(min, max uint) Quantifier {
	return Quantifier{min: min, max: max}
}

// Count returns a new Quantifier with exact n runes to match.
func Count(n uint) Quantifier {
	return Quantifier{min: n, max: n}
}

// String implements fmt.Stringer interface.
func (q Quantifier) String() string {
	return fmt.Sprintf("min: %d, max: %d", q.min, q.max)
}

func (q Quantifier) isValid() (ret bool) {
	ret = q.min <= q.max
	return
}

func (q Quantifier) isZero() (ret bool) {
	ret = q.min == q.max && q.min == 0
	return
}

func (q Quantifier) isOne() (ret bool) {
	ret = q.min == q.max && q.min == 1
	return
}

func (q Quantifier) makeResult(repeats uint) (err error) {
	switch {
	case q.min == q.max:
		if repeats != q.min {
			err = ErrRollback
		} else {
			err = ErrNext
		}
	case q.min < q.max:
		if repeats < q.min || repeats > q.max {
			err = ErrRollback
		} else {
			err = ErrNext
		}
	case q.min == 0:
		if repeats > q.max {
			err = ErrRollback
		} else {
			err = ErrNext
		}
	case q.max == math.MaxUint:
		if repeats < q.min {
			err = ErrRollback
		} else {
			err = ErrNext
		}
	default:
		panic("unreachable")
	}
	return
}

type Repeat[T any] struct {
	logger common.Logger
	q      Quantifier
}

func newRepeat[T any](logger common.Logger, q Quantifier) *Repeat[T] {
	return &Repeat[T]{
		logger: logger,
		q:      q,
	}
}

func (r Repeat[T]) Update(ctx context.Context, tx xio.State) (err error) {
	switch {
	case r.q.isZero():
		err = ErrRollback
	case r.q.isOne():
		err = ErrNext
	default:
		err = MakeRepeat(r.q)
	}
	return
}

func isRepeat[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Repeat[T])
	return
}

func isZeroMinRepeat[T any](s Update[T]) (ret bool) {
	repeat, ok := s.(*Repeat[T])
	if !ok {
		return
	}
	ret = repeat.q.min == 0
	return
}

func isZeroMaxRepeat[T any](s Update[T]) (ret bool) {
	repeat, ok := s.(*Repeat[T])
	if !ok {
		return
	}
	ret = repeat.q.max == 0
	return
}

func isRepeatable[T any](s Update[T]) (ret bool) {
	if isRepeat[T](s) {
		return
	}
	if isEmit[T](s) {
		return
	}
	if isError[T](s) {
		return
	}
	if isOmit[T](s) {
		return
	}
	return true
}

// repeat implements repeat sub state.
func (c *Chain[T]) repeat(ctx context.Context, state Update[T], repeat error, tx xio.State) (err error) {
	if state == nil {
		c.logger.Fatal("invalid grammar: repeat without previous state")
	}
	q, ok := getQuantifier(repeat)
	if !ok {
		c.logger.Fatal("not a quantifier: %s", repeat)
	}
	if q.max == 1 {
		err = ErrNext
		return
	}
	source := xio.AsSource(tx)
	count := uint(1)
loop:
	for ; count < q.max; count++ {
		tx := source.Begin()
		if err = state.Update(ctx, tx); err == nil {
			c.logger.Fatal("unexpected nil")
		}
		switch {
		case errors.Is(err, ErrRollback):
			if err := tx.Rollback(); err != nil {
				c.logger.Fatal("rollback error: %s", err)
			}
			err = q.makeResult(count)
			break loop
		case errors.Is(err, ErrNext), errors.Is(err, ErrCommit):
			if err := tx.Commit(); err != nil {
				c.logger.Fatal("commit error: %s", err)
			}
			nextCount := count + 1
			if nextCount < q.max {
				continue
			}
			err = q.makeResult(nextCount)
			break loop
		default:
			if err := tx.Rollback(); err != nil {
				c.logger.Fatal("rollback error: %s", err)
			}
			c.logger.Fatal("unexpected error: %s", err)
		}
	}
	return
}

func (b Builder[T]) repeat(defaultName string, q Quantifier) (tail *Chain[T]) {
	if !q.isValid() {
		b.logger.Fatal("invalid grammar: invalid quantifier: %s", q)
	}
	if b.last == nil {
		b.logger.Fatal("invalid grammar: repeat can't be the first state in chain")
	}
	if !isRepeatable[T](b.last.state) {
		b.logger.Fatal("invalid grammar: previous state '%s' is not repeatable", b.last.name)
	}
	tail = b.createNode(defaultName, func() any { return newRepeat[T](b.logger, q) })
	return
}

// Repeat is a quantifier for previous state.
func (b Builder[T]) Repeat(q Quantifier) (tail *Chain[T]) {
	tail = b.repeat("Repeat", q)
	return
}

// Optional is a quantifier for previous state.
func (b Builder[T]) Optional() (tail *Chain[T]) {
	tail = b.repeat("Optional", CountBetween(0, 1))
	return
}
