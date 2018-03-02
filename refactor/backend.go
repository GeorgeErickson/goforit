package refactor

import (
	"sync"
	"time"
)

// A Backend knows the current set of flags
type Backend interface {
	// Flags gets the flag with the given name.
	// Also yields the last time flags were updated (zero time if unknown)
	Flag(name string) (Flag, time.Time, error)

	// SetErrorHandler adds a handler that will be called if this backend encounters
	// errors.
	SetErrorHandler(ErrorHandler)

	// SetAgeCallback adds a handler that should be called whenever flags are updated.
	SetAgeCallback(AgeCallback)

	// Close releases any resources held by this backend
	Close() error
}

type BackendBase struct {
	handlerMtx   sync.RWMutex
	errorHandler ErrorHandler
	ageCallback  AgeCallback
}

func (b *BackendBase) SetErrorHandler(h ErrorHandler) {
	b.handlerMtx.Lock()
	defer b.handlerMtx.Unlock()
	b.errorHandler = h
}

func (b *BackendBase) handleError(err error) error {
	b.handlerMtx.RLock()
	defer b.handlerMtx.RUnlock()
	if b.errorHandler != nil {
		go b.errorHandler(err)
	}
	return nil
}

func (b *BackendBase) SetAgeCallback(cb AgeCallback) {
	b.handlerMtx.Lock()
	defer b.handlerMtx.Unlock()
	b.ageCallback = cb
}

func (b *BackendBase) handleAge(age time.Duration) {
	b.handlerMtx.RLock()
	defer b.handlerMtx.RUnlock()
	if b.ageCallback != nil {
		go b.ageCallback(AgeSource, age)
	}
}

func (b *BackendBase) Close() error {
	return nil
}

// OffBackend is a backend where every flag is off.
// Good for just turning off all logging.
type OffBackend struct {
	BackendBase
}

func (*OffBackend) Flag(name string) (Flag, time.Time, error) {
	return SampleFlag{FlagName: name, Rate: 0}, time.Time{}, nil
}
