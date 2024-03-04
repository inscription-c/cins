package util

import "testing"

func TestNewPeekable(t *testing.T) {
	s := []int{1, 2, 3}
	peek, err := NewPeekable(&s)
	if err != nil {
		t.Fatal(err)
	}
	if peek.Peek().(int) != 1 {
		t.Fatal("peek error")
	}
	if peek.Next().(int) != 1 {
		t.Fatal("next error")
	}
	if peek.Peek().(int) != 2 {
		t.Fatal("peek error")
	}
	if peek.Next().(int) != 2 {
		t.Fatal("next error")
	}
	if peek.Peek().(int) != 3 {
		t.Fatal("peek error")
	}
	if peek.Next().(int) != 3 {
		t.Fatal("next error")
	}
	if peek.Peek() != nil {
		t.Fatal("peek error")
	}
	if peek.Next() != nil {
		t.Fatal("next error")
	}
	if peek.HasNext() {
		t.Fatal("has next error")
	}
}
