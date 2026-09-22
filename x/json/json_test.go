package json

import "testing"

func TestMarshal(t *testing.T) {
	got := Marshal(map[string]int{"a": 1})
	if string(got) != `{"a":1}` {
		t.Fatalf("Marshal() = %s, want %s", got, `{"a":1}`)
	}
}

func TestMarshal_UnmarshalableValueReturnsNilNotPanic(t *testing.T) {
	got := Marshal(make(chan int))
	if got != nil {
		t.Fatalf("Marshal(chan) = %s, want nil on a marshal error", got)
	}
}

func TestMarshalIndent(t *testing.T) {
	got := MarshalIndent(map[string]int{"a": 1}, "", "  ")
	want := "{\n  \"a\": 1\n}"
	if string(got) != want {
		t.Fatalf("MarshalIndent() = %s, want %s", got, want)
	}
}

func TestUnmarshal(t *testing.T) {
	var m map[string]int
	got := Unmarshal([]byte(`{"a":1}`), &m)

	gotMap, ok := got.(*map[string]int)
	if !ok {
		t.Fatalf("Unmarshal() returned %T, want *map[string]int", got)
	}
	if (*gotMap)["a"] != 1 {
		t.Fatalf("Unmarshal() = %v, want map[a:1]", *gotMap)
	}
}

func TestUnmarshal_InvalidJSONReturnsTargetUnchanged(t *testing.T) {
	target := &struct{ A int }{A: 5}
	got := Unmarshal([]byte("not json"), target)
	if got != target {
		t.Fatalf("Unmarshal() = %v, want the same target pointer back", got)
	}
	if target.A != 5 {
		t.Fatalf("target.A = %d, want unchanged 5 after an unmarshal error", target.A)
	}
}
