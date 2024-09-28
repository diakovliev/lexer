package state

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	// Bytes is a state that matches the given bytes.
	Bytes struct {
		logger   common.Logger
		provider BytesSamplesProvider
		pred     bytesPredicate
	}

	// BytesSamplesProvider is a function that returns the slice of a sample bytes to match.
	BytesSamplesProvider func() [][]byte

	// StringSamplesProvider is a function that returns the slice of a sample strings to match.
	StringSamplesProvider func() []string

	// bytesPredicate is a function that checks if the given bytes matches the samples.
	bytesPredicate func(in []byte, samples [][]byte) (int, bool)
)

func newBytes[T any](
	logger common.Logger,
	provider BytesSamplesProvider,
	pred bytesPredicate,
) *Bytes {
	return &Bytes{
		logger:   logger,
		provider: provider,
		pred:     pred,
	}
}

// Update implements State interface.
func (bs Bytes) Update(ctx context.Context, tx xio.State) (err error) {
	samples := bs.provider()
	maxLen := 0
	for _, sample := range samples {
		l := len(sample)
		common.AssertFalse(l == 0, "invalid grammar: empty sample")
		if len(sample) > maxLen {
			maxLen = len(sample)
		}
	}
	common.AssertFalse(maxLen == 0, "invalid grammar: max sample len is zero")
	buffer := make([]byte, maxLen)
	n, err := tx.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	in := buffer[:n]
	matched, ok := bs.pred(in, samples)
	if !ok {
		_, err = tx.Unread()
		common.AssertNoError(err, "unread error")
		err = ErrRollback
		return
	}
	// adjust tx position
	_, err = tx.Unread()
	common.AssertNoError(err, "unread error")
	n, err = tx.Read(buffer[:matched])
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	common.AssertTrue(n == matched, "unexpected read length")
	err = ErrChainNext
	return
}

func providerFromBytes(samples [][]byte) BytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range samples {
			common.AssertFalse(len(s) == 0, "invalid grammar: empty sample")
			ret = append(ret, []byte(s))
		}
		return
	}
}

func providerFromStrings(samples []string) BytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range samples {
			common.AssertFalse(len(s) == 0, "invalid grammar: empty sample")
			ret = append(ret, []byte(s))
		}
		return
	}
}

func bytesMatches(in []byte, samples [][]byte) (n int, ret bool) {
	var maxLen int
	for _, sample := range samples {
		n = len(sample)
		if bytes.Equal(sample, in[:n]) {
			ret = true
			return
		}
		if n > maxLen {
			maxLen = n
		}
	}
	return maxLen, false
}

func bytesNotMatches(in []byte, samples [][]byte) (n int, ret bool) {
	n, ret = bytesMatches(in, samples)
	ret = !ret
	return
}

func asBytesProvider(sp StringSamplesProvider) BytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range sp() {
			ret = append(ret, []byte(s))
		}
		return
	}
}

func (b Builder[T]) bytesState(name string, provider BytesSamplesProvider, pred bytesPredicate) (tail *Chain[T]) {
	tail = b.append(name, func() Update[T] { return newBytes[T](b.logger, provider, pred) })
	return
}

// BytesFn matches any sample from given samples.
func (b Builder[T]) BytesFn(fn BytesSamplesProvider) (tail *Chain[T]) {
	tail = b.bytesState("BytesFn", fn, bytesMatches)
	return
}

// NotBytesFn matches any byte sequence with maximum sample len except for the given samples.
func (b Builder[T]) NotBytesFn(fn BytesSamplesProvider) (tail *Chain[T]) {
	tail = b.bytesState("NotBytesFn", fn, bytesNotMatches)
	return
}

// Bytes matches any sample from given samples.
func (b Builder[T]) Bytes(samples ...[]byte) (tail *Chain[T]) {
	tail = b.bytesState("Bytes", providerFromBytes(samples), bytesMatches)
	return
}

// BytesNot matches any byte sequence with maximum sample len except for the given samples.
func (b Builder[T]) NotBytes(samples ...[]byte) (tail *Chain[T]) {
	tail = b.bytesState("NotBytes", providerFromBytes(samples), bytesNotMatches)
	return
}

// StringFn matches any sample from given samples.
func (b Builder[T]) StringFn(fn StringSamplesProvider) (tail *Chain[T]) {
	tail = b.bytesState("StringFn", asBytesProvider(fn), bytesMatches)
	return
}

// NotStringFn matches any byte sequence with maximum sample len except for the given samples.
func (b Builder[T]) NotStringFn(fn StringSamplesProvider) (tail *Chain[T]) {
	tail = b.bytesState("NotStringFn", asBytesProvider(fn), bytesNotMatches)
	return
}

// String matches any sample from given samples.
func (b Builder[T]) String(samples ...string) (tail *Chain[T]) {
	tail = b.bytesState("String", providerFromStrings(samples), bytesMatches)
	return
}

// NotString matches any string with maximum sample len except for the given samples.
func (b Builder[T]) NotString(samples ...string) (tail *Chain[T]) {
	tail = b.bytesState("NotString", providerFromStrings(samples), bytesNotMatches)
	return
}
