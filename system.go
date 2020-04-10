package fam

import (
	"reflect"

	"github.com/jakecoffman/eng"
)

// System uses reflection to implement common functionality. The hot paths
// are update and draw and they aren't implemented here so they're still fast.
type System struct {
	active    int
	pool      interface{}
	maxAmount int
	lookup    map[eng.EntityID]int

	typ reflect.Type
}

func NewSystem(obj interface{}, maxAmount int) *System {
	t := reflect.TypeOf(obj)
	return &System{
		pool:      reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(obj)), maxAmount, maxAmount).Interface(),
		maxAmount: maxAmount,
		lookup:    map[eng.EntityID]int{},
		typ:       t,
	}
}

func (s *System) Add() (interface{}, bool) {
	poolSlice := reflect.ValueOf(s.pool)
	if s.active >= s.maxAmount {
		// we've reached our max so instead of crashing return the last one
		item := poolSlice.Index(s.active-1)
		return item.Addr().Interface(), false
	}
	item := poolSlice.Index(s.active)
	id := eng.NextEntityID()
	item.FieldByName("ID").Set(reflect.ValueOf(id))
	p := item.Addr().Interface()
	s.lookup[id] = s.active
	s.active++
	return p, true
}

func (s *System) Get(id eng.EntityID) interface{} {
	return reflect.ValueOf(s.pool).Index(s.lookup[id]).Addr().Interface()
}

func (s *System) Remove(id eng.EntityID) {
	index, ok := s.lookup[id]
	if !ok {
		return
	}
	delete(s.lookup, id)
	s.active--
	poolSlice := reflect.ValueOf(s.pool)
	itemRemove := poolSlice.Index(index)
	itemKeep := poolSlice.Index(s.active)
	s.lookup[eng.EntityID(itemKeep.FieldByName("ID").Int())] = index
	itemRemove.Set(itemKeep)
	return
}
