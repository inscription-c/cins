package util

import "errors"

import (
	"reflect"
	"sync"
)

// ErrNotSlice is an error that is returned when the provided value is not a slice.
var ErrNotSlice = errors.New("not slice")

// ErrSliceValueCannotInterface is an error that is returned when the slice value cannot be converted to an interface.
var ErrSliceValueCannotInterface = errors.New("slice value cannot interface")

// Peekable is an interface that defines methods for peeking and iterating over a slice.
type Peekable interface {
	Peek() interface{}                           // Peek returns the current element in the slice without moving the iterator.
	Next() interface{}                           // Next returns the current element and moves the iterator to the next element.
	HasNext() bool                               // HasNext checks if there are more elements in the slice.
	Range(fn func(int, interface{}) error) error // Range iterates over the slice and calls the provided function for each element.
}

// Peek is a struct that implements the Peekable interface. It contains a slice and an index for iteration.
type Peek struct {
	l     sync.RWMutex  // l is a read-write lock for safe concurrent access.
	idx   int           // idx is the current index in the slice.
	slice []interface{} // slice is the slice of elements.
}

// Peek returns the current element in the slice without moving the iterator.
func (p *Peek) Peek() interface{} {
	p.l.RLock()
	defer p.l.RUnlock()
	if p.idx > len(p.slice)-1 {
		return nil
	} else {
		return p.slice[p.idx]
	}
}

// Next returns the current element and moves the iterator to the next element.
func (p *Peek) Next() interface{} {
	v := p.Peek()
	if v == nil {
		return nil
	}
	p.l.Lock()
	p.idx++
	p.l.Unlock()
	return v
}

// HasNext checks if there are more elements in the slice.
func (p *Peek) HasNext() bool {
	p.l.RLock()
	defer p.l.RUnlock()
	if p.idx > len(p.slice)-1 {
		return false
	} else {
		return true
	}
}

// Range iterates over the slice and calls the provided function for each element.
func (p *Peek) Range(fn func(idx int, v interface{}) error) error {
	p.l.RLock()
	defer p.l.RUnlock()

	for v := p.Peek(); ; v = p.Next() {
		if v == nil {
			break
		}
		if err := fn(p.idx, v); err != nil {
			return err
		}
	}
	return nil
}

// NewPeekable creates a new Peekable from a slice or a pointer to a slice.
// It returns an error if the provided value is not a slice or if the slice values cannot be converted to an interface.
func NewPeekable(slice interface{}) (Peekable, error) {
	rv := reflect.ValueOf(slice)
	rk := rv.Type().Kind()
	if rk != reflect.Slice && rk != reflect.Ptr {
		return nil, ErrNotSlice
	}
	if rk == reflect.Ptr {
		rv = rv.Elem()
		k := rv.Type().Kind()
		if k != reflect.Slice {
			return nil, ErrNotSlice
		}
	}

	p := &Peek{
		slice: make([]interface{}, rv.Len()),
	}
	for i := 0; i < rv.Len(); i++ {
		if !rv.Index(i).CanInterface() {
			return nil, ErrSliceValueCannotInterface
		}
		p.slice[i] = rv.Index(i).Interface()
	}
	return p, nil
}
