package states

import (
	"errors"

	"github.com/diakovliev/lexer/common"
)

type StateProvider[T any] func(b Builder[T]) []State[T]

type SubState[T any] struct {
	logger   common.Logger
	builder  Builder[T]
	provider StateProvider[T]
	states   []State[T]
	current  int
}

func newSubState[T any](logger common.Logger, states ...State[T]) *SubState[T] {
	return &SubState[T]{
		logger: logger,
		states: states,
	}
}

func newSubStateProvider[T any](logger common.Logger, builder Builder[T], provider StateProvider[T]) *SubState[T] {
	return &SubState[T]{
		logger:   logger,
		provider: provider,
		builder:  builder,
	}
}

// currentState returns the current state of the lexer.
func (ss *SubState[T]) currentState() State[T] {
	if len(ss.states) == 0 && ss.provider != nil {
		ss.states = ss.provider(ss.builder)
	}
	if len(ss.states) == 0 {
		return nil
	}
	if len(ss.states) <= ss.current {
		return nil
	}
	return ss.states[ss.current]
}

// next moves the lexer to the next state.
func (ss *SubState[T]) next() {
	ss.current++
}

// reset resets the lexer to its first state.
func (ss *SubState[T]) reset() {
	ss.current = 0
}

// update updates the current state of the lexer with the given transaction.
func (ss *SubState[T]) update(tx common.ReadUnreadData) (err error) {
	state := ss.currentState()
	if state == nil {
		// no more states to process, we're done
		err = ErrNoMoreStates
		return
	}
	err = state.Update(tx)
	return
}

// we don't want to expose common.Tx to the states implementations,
// so we'll use this helper go get the tx from the ReadUnreadData interface here in Update.
func (ss SubState[T]) asTx(rud common.ReadUnreadData) (tx common.Tx) {
	var i any = rud
	tx, ok := i.(common.Tx)
	if !ok {
		ss.logger.Fatal("not a common.Tx")
	}
	return
}

// Update implements State interface. It updates the current state of the lexer with the given transaction.
func (ss *SubState[T]) Update(parentTx common.ReadUnreadData) (err error) {
	ss.logger.Trace("=>> enter SubState.Update()")
	defer func() { ss.logger.Trace("<<= leave SubState.Update() = err=%s", err) }()

loop:
	for {
		tx := ss.asTx(parentTx).Begin()
		if err = ss.update(tx); err == nil {
			ss.logger.Fatal("unexpected nil")
		}
		switch {
		case errors.Is(err, ErrCommit):
			ss.logger.Trace("ErrCommit")
			if _, commitErr := tx.Commit(); commitErr != nil {
				ss.logger.Error("ErrCommit -> commit error: %v", commitErr)
				err = commitErr
				break loop
			}
			ss.reset()
		case errors.Is(err, ErrRollback):
			ss.logger.Trace("ErrRollback")
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				ss.logger.Error("ErrRollback -> rollback error: %v", rollbackErr)
				err = rollbackErr
				break loop
			}
			ss.next()
		case errors.Is(err, ErrNoMoreStates):
			ss.logger.Trace("ErrNoMoreStates")
			if tx.Has() {
				ss.logger.Error("has non processed data")
				err = ErrHasMoreData
			} else {
				ss.logger.Error("incomplete state")
				err = ErrIncompleteSubState
			}
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				ss.logger.Error("ErrNoMoreStates -> rollback error: %v", rollbackErr)
				err = rollbackErr
				break loop
			}
			break loop
		case errors.Is(err, errBreak):
			ss.logger.Trace("break")
			if _, commitErr := tx.Commit(); commitErr != nil {
				ss.logger.Error("ErrCommit -> commit error: %v", commitErr)
				err = commitErr
				break loop
			}
			err = ErrCommit
			break loop
		default:
			ss.logger.Error("unexpected error: %v", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				ss.logger.Error("%s -> rollback error: %v", err, rollbackErr)
				err = rollbackErr
			}
			break loop
		}
	}
	return
}

func (b Builder[T]) State(states ...State[T]) (head *Chain[T]) {
	defaultName := "SubState"
	head = b.createNode(defaultName, func() any { return newSubState(b.logger, states...) })
	return
}

func (b Builder[T]) StateProvider(builder Builder[T], provider StateProvider[T]) (head *Chain[T]) {
	defaultName := "SubStateProvider"
	head = b.createNode(defaultName, func() any { return newSubStateProvider(b.logger, builder, provider) })
	return
}
