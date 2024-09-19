package states

import (
	"errors"
	"io"
	"math"

	"github.com/diakovliev/lexer/common"
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

func (q Quantifier) isZero() (ret bool) {
	ret = q.min == q.max && q.min == 0
	return
}

func (q Quantifier) inRange(count int) (ret bool) {
	ret = q.min <= count && (q.max == math.MaxInt || count <= q.max)
	return
}

func (q Quantifier) getResult(count int) (err error) {
	switch {
	case q.min == q.max:
		if count != q.min {
			err = ErrRollback
		} else {
			err = errChainNext
		}
	case q.min < q.max:
		if count < q.min || count > q.max {
			err = ErrRollback
		} else {
			err = errChainNext
		}
	case q.min == 0:
		if count < 0 || count > q.max {
			err = ErrRollback
		} else {
			err = errChainNext
		}
	case q.max == math.MaxInt:
		if count < q.min {
			err = ErrRollback
		} else {
			err = errChainNext
		}
	default:
		panic("unreachable")
	}
	return
}

// Quantified matches runes by given function using given quantifier.
type Quantified[T any] struct {
	logger common.Logger
	fn     func(rune) bool
	q      Quantifier
}

func newQuantified[T any](logger common.Logger, fn func(rune) bool, q Quantifier) *Quantified[T] {
	return &Quantified[T]{
		logger: logger,
		fn:     fn,
		q:      q,
	}
}

func (qq Quantified[T]) Update(tx common.ReadUnreadData) (err error) {
	if qq.q.isZero() {
		err = errChainNext
		return
	}
	count := 0
	for !qq.q.inRange(count) {
		data, r, nextErr := common.NextRuneFrom(tx)
		if nextErr != nil && !errors.Is(nextErr, io.EOF) {
			err = nextErr
			return
		}
		if errors.Is(nextErr, io.EOF) && len(data) == 0 {
			break
		}
		if !qq.fn(r) {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				qq.logger.Error("unread error: %s", unreadErr)
			}
			break
		}
		count++
	}
	err = qq.q.getResult(count)
	return
}

// QuantifiedFn matches runes by given function using given quantifier.
func (b Builder[T]) QuantifiedFn(fn func(rune) bool, q Quantifier) (head *Chain[T]) {
	defaultName := "QuantifiedFn"
	head = b.createNode(defaultName, func() any { return newQuantified[T](b.logger, fn, q) })
	return
}

// QuantifiedRune matches rune by given function using given quantifier.
func (b Builder[T]) QuantifiedRune(ir rune, q Quantifier) (head *Chain[T]) {
	defaultName := "QuantifiedRune"
	head = b.createNode(defaultName, func() any { return newQuantified[T](b.logger, func(r rune) bool { return r == ir }, q) })
	return
}
