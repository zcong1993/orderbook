package main

import (
	"reflect"
	"testing"
)

func Test_Add(t *testing.T) {
	expect(t, Add(1, 2), 3)
	expect(t, Add(1, 3), 4)
}

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
