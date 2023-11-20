// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package list is a generic and thus type-safe version of golang's container/list.
// Wherever possible, it mimics the exact internal behavior as closely as possible.
package list

// Element is an element of a List.
type Element[E any] struct {
	next *Element[E]
	prev *Element[E]

	list *List[E]

	// Value is the value of the list Element.
	Value E
}

// Next returns the next Element in the List, if any.
func (e *Element[E]) Next() *Element[E] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous Element in the List, if any.
func (e *Element[E]) Prev() *Element[E] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// List is a doubly-linked list.
type List[E any] struct {
	root Element[E]
	len  int
}

// New constructs a new empty List.
func New[E any]() *List[E] {
	return new(List[E]).Init()
}

// Init initializes the list and pointers of the list.
func (l *List[E]) Init() *List[E] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// Len returns the number of Elements in the list.
func (l *List[E]) Len() int {
	return l.len
}

// Front returns the Element at the front of the list, if any.
func (l *List[E]) Front() *Element[E] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the Element at the back of the list, if any.
func (l *List[E]) Back() *Element[E] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *List[E]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

func (l *List[E]) insert(e, at *Element[E]) *Element[E] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

func (l *List[E]) insertValue(e E, at *Element[E]) *Element[E] {
	return l.insert(&Element[E]{Value: e}, at)
}

func (l *List[E]) remove(e *Element[E]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil
	l.len--
}

func (l *List[E]) move(e, at *Element[E]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev
	e.prev.next = e
	e.next.prev = e
}

// Remove removes the Element from the list if it was part of the List.
func (l *List[E]) Remove(e *Element[E]) E {
	if e.list == l {
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts the value at the front of the list, returning the created Element.
func (l *List[E]) PushFront(e E) *Element[E] {
	l.lazyInit()
	return l.insertValue(e, &l.root)
}

// PushBack inserts the value at the back of the list, returning the created Element.
func (l *List[E]) PushBack(e E) *Element[E] {
	l.lazyInit()
	return l.insertValue(e, l.root.prev)
}

// InsertBefore inserts the value before the Element used as mark, if the Element is part of the list.
func (l *List[E]) InsertBefore(e E, mark *Element[E]) *Element[E] {
	if mark.list != l {
		return nil
	}
	return l.insertValue(e, mark.prev)
}

// InsertAfter inserts the value after the Element used as mark, if the Element is part of the list.
func (l *List[E]) InsertAfter(e E, mark *Element[E]) *Element[E] {
	if mark.list != l {
		return nil
	}
	// see comment in List[E].Remove about initialization of l
	return l.insertValue(e, mark)
}

// MoveToFront moves the Element to the front of the List if it is part of it.
func (l *List[E]) MoveToFront(e *Element[E]) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List[E].Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves the Element to the back of the List if it is part of it.
func (l *List[E]) MoveToBack(e *Element[E]) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List[E].Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves the Element before the marked element if it is part of the List.
func (l *List[E]) MoveBefore(e, mark *Element[E]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves the Element after the marked element if it is part of the List.
func (l *List[E]) MoveAfter(e, mark *Element[E]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList pushes all values of the other list to the back of this List.
func (l *List[E]) PushBackList(other *List[E]) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList pushes all values of the other list to the front of this List.
func (l *List[E]) PushFrontList(other *List[E]) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
