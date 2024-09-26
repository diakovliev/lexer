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
		provider bytesSamplesProvider
		pred     bytesPredicate
	}

	// bytesSamplesProvider is a function that returns the slice of a sample bytes to match.
	bytesSamplesProvider func() [][]byte

	// bytesPredicate is a function that checks if the given bytes matches the samples.
	bytesPredicate func(in []byte, samples [][]byte) bool
)

func newBytes[T any](
	logger common.Logger,
	provider bytesSamplesProvider,
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
	if !bs.pred(in, samples) {
		_, err = tx.Unread()
		common.AssertNoError(err, "unread error")
		err = ErrRollback
		return
	}
	err = ErrChainNext
	return
}

func providerFromBytes(samples [][]byte) bytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range samples {
			common.AssertFalse(len(s) == 0, "invalid grammar: empty sample")
			ret = append(ret, []byte(s))
		}
		return
	}
}

func providerFromStrings(samples []string) bytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range samples {
			common.AssertFalse(len(s) == 0, "invalid grammar: empty sample")
			ret = append(ret, []byte(s))
		}
		return
	}
}

func bytesMatches(in []byte, samples [][]byte) bool {
	for _, sample := range samples {
		if bytes.Equal(sample, in) {
			return true
		}
	}
	return false
}

func bytesNotMatches(in []byte, samples [][]byte) bool {
	return !bytesMatches(in, samples)
}

func (b Builder[T]) bytesState(name string, provider bytesSamplesProvider, pred bytesPredicate) (tail *Chain[T]) {
	tail = b.append(name, func() Update[T] { return newBytes[T](b.logger, provider, pred) })
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
