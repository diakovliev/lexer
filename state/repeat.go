package state

import (
	"context"
	"fmt"
	"math"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Quantifier is a Count state quantifier.
// Cases:
//   - min == max: match exact count of runes
//   - min < max: match any count of runes in range [min, max]
//   - min == 0: match any count of runes in range [0, max]
//   - max == math.MaxInt: match any count of runes in range [min, infinity(EOF)]
type Quantifier struct {
	min int
	max int
}

// CountBetween returns a new Quantified state quantifier with min and max runes to match.
func CountBetween(min, max int) Quantifier {
	return Quantifier{min: min, max: max}
}

// Count returns a new Quantified state quantifier with exact n runes to match.
func Count(n int) Quantifier {
	return Quantifier{min: n, max: n}
}

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

func (q Quantifier) InRange(repeats int) (ret bool) {
	ret = q.min <= repeats && (q.max == math.MaxInt || repeats <= q.max)
	return
}

func (q Quantifier) MakeResult(repeats int) (err error) {
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
		if repeats < 0 || repeats > q.max {
			err = ErrRollback
		} else {
			err = ErrNext
		}
	case q.max == math.MaxInt:
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

// Repeat is a state that applies a	quantifier to a previous state.
func (b Builder[T]) Repeat(q Quantifier) (head *Chain[T]) {
	if !q.isValid() {
		b.logger.Fatal("invalid grammar: invalid quantifier: %s", q)
	}
	defaultName := "Q"
	head = b.createNode(defaultName, func() any { return newRepeat[T](b.logger, q) })
	return
}
