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
		pred     BytesPredicate
	}

	// BytesSamplesProvider is a function that returns the slice of a sample bytes to match.
	BytesSamplesProvider func() [][]byte
	// BytesPredicate is a function that checks if the given bytes matches the samples.
	BytesPredicate func(in []byte, samples [][]byte) bool
)

func newBytes[T any](
	logger common.Logger,
	provider BytesSamplesProvider,
	pred BytesPredicate,
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
		switch {
		case l == 0:
			bs.logger.Fatal("invalid grammar: empty sample")
		case len(sample) > maxLen:
			maxLen = len(sample)
		}
	}
	if maxLen == 0 {
		bs.logger.Fatal("max sample len is zero")
	}
	buffer := make([]byte, maxLen)
	n, err := tx.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	in := buffer[:n]
	if !bs.pred(in, samples) {
		err = errRollback
		return
	}
	err = errNext
	return
}

func providerFromBytes(logger common.Logger, samples [][]byte) BytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range samples {
			if len(s) == 0 {
				logger.Fatal("invalid grammar: empty sample")
			}
			ret = append(ret, []byte(s))
		}
		return
	}
}

func providerFromStrings(logger common.Logger, samples []string) BytesSamplesProvider {
	return func() (ret [][]byte) {
		for _, s := range samples {
			if len(s) == 0 {
				logger.Fatal("invalid grammar: empty sample")
			}
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

func (b Builder[T]) bytesState(name string, provider BytesSamplesProvider, pred BytesPredicate) (tail *Chain[T]) {
	tail = b.append(name, func() any { return newBytes[T](b.logger, provider, pred) })
	return
}

func (b Builder[T]) Bytes(samples ...[]byte) (tail *Chain[T]) {
	tail = b.bytesState("Bytes", providerFromBytes(b.logger, samples), bytesMatches)
	return
}

func (b Builder[T]) NotBytes(samples ...[]byte) (tail *Chain[T]) {
	tail = b.bytesState("NotBytes", providerFromBytes(b.logger, samples), bytesNotMatches)
	return
}

// String is a state that compares the given samples with state input.
// It will has positive result if any sample is equal to state input.
func (b Builder[T]) String(samples ...string) (tail *Chain[T]) {
	tail = b.bytesState("String", providerFromStrings(b.logger, samples), bytesMatches)
	return
}

// NotString is a state that compares the given samples with state input.
// It will has positive result if nothing from samples is equal to state input.
func (b Builder[T]) NotString(samples ...string) (tail *Chain[T]) {
	tail = b.bytesState("NotString", providerFromStrings(b.logger, samples), bytesNotMatches)
	return
}
