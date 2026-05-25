package sitediff

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func mkArr(t *testing.T, raw string) []any {
	t.Helper()
	var out []any
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

func TestMatchArrays_Columns_ByColumnName(t *testing.T) {
	old := mkArr(t, `[{"ColumnName":"A"},{"ColumnName":"B"},{"ColumnName":"C"}]`)
	newA := mkArr(t, `[{"ColumnName":"B"},{"ColumnName":"D"},{"ColumnName":"A"}]`) // reorder + B kept, C removed, D added
	matched, onlyOld, onlyNew := MatchArrays("Columns", old, newA)

	// A:0->2, B:1->0
	if len(matched) != 2 {
		t.Fatalf("matched count: %d (%+v)", len(matched), matched)
	}
	pairs := map[int]int{}
	for _, m := range matched {
		pairs[m.OldIndex] = m.NewIndex
	}
	if pairs[0] != 2 || pairs[1] != 0 {
		t.Errorf("matched pairs wrong: %v", pairs)
	}
	if !reflect.DeepEqual(onlyOld, []int{2}) {
		t.Errorf("onlyOld = %v, want [2]", onlyOld)
	}
	if !reflect.DeepEqual(onlyNew, []int{1}) {
		t.Errorf("onlyNew = %v, want [1]", onlyNew)
	}
}

func TestMatchArrays_Scripts_TitleThenId(t *testing.T) {
	old := mkArr(t, `[{"Id":1,"Title":"hdr","Body":"x"},{"Id":2,"Title":"foo","Body":"y"}]`)
	// Title preserved on idx 0, idx 1 has Title removed but Id preserved
	newA := mkArr(t, `[{"Id":1,"Title":"hdr","Body":"x"},{"Id":2,"Body":"z"}]`)
	matched, _, _ := MatchArrays("Scripts", old, newA)
	if len(matched) != 2 {
		t.Fatalf("expected both scripts to match, got %d (%+v)", len(matched), matched)
	}
}

func TestMatchArrays_Scripts_HashFallback(t *testing.T) {
	// Both old and new have empty Title and no Id, but the Body is identical
	// — the body-hash fallback must still match them.
	old := mkArr(t, `[{"Body":"console.log('hello world');"}]`)
	newA := mkArr(t, `[{"Body":"console.log('hello world');"}]`)
	matched, _, _ := MatchArrays("Scripts", old, newA)
	if len(matched) != 1 {
		t.Fatalf("hash fallback failed: matched=%v", matched)
	}
}

func TestMatchArrays_Notifications_CompositeKey(t *testing.T) {
	old := mkArr(t, `[{"Type":"Mail","Address":"a@x"},{"Type":"Mail","Address":"b@x"}]`)
	newA := mkArr(t, `[{"Type":"Mail","Address":"b@x"},{"Type":"Mail","Address":"a@x"}]`)
	matched, onlyOld, onlyNew := MatchArrays("Notifications", old, newA)
	if len(matched) != 2 || len(onlyOld) != 0 || len(onlyNew) != 0 {
		t.Fatalf("expected full match, got matched=%d onlyOld=%v onlyNew=%v", len(matched), onlyOld, onlyNew)
	}
}

func TestMatchArrays_GridColumns_OrderedByIndex(t *testing.T) {
	// GridColumns is Ordered=true; same content but reordered → diff at both positions.
	old := mkArr(t, `["A","B"]`)
	newA := mkArr(t, `["B","A"]`)
	matched, onlyOld, onlyNew := MatchArrays("GridColumns", old, newA)
	if len(matched) != 2 || len(onlyOld) != 0 || len(onlyNew) != 0 {
		t.Fatalf("ordered: expected 2 indexed pairs, got %d", len(matched))
	}
	for _, m := range matched {
		if m.OldIndex != m.NewIndex {
			t.Errorf("ordered pair must align indices: %+v", m)
		}
	}
}

func TestMatchArrays_UnknownArray_FallsBackToIndex(t *testing.T) {
	old := mkArr(t, `[{"X":1},{"X":2}]`)
	newA := mkArr(t, `[{"X":1},{"X":2},{"X":3}]`)
	matched, onlyOld, onlyNew := MatchArrays("WhateverNew", old, newA)
	if len(matched) != 2 || len(onlyOld) != 0 || len(onlyNew) != 1 {
		t.Fatalf("index fallback wrong: matched=%d onlyOld=%v onlyNew=%v", len(matched), onlyOld, onlyNew)
	}
}

func TestMatchArrays_DuplicateKeys_FirstWinsRestAreNew(t *testing.T) {
	old := mkArr(t, `[{"ColumnName":"A"}]`)
	newA := mkArr(t, `[{"ColumnName":"A"},{"ColumnName":"A"}]`)
	matched, _, onlyNew := MatchArrays("Columns", old, newA)
	if len(matched) != 1 {
		t.Fatalf("expected exactly one match, got %d", len(matched))
	}
	if len(onlyNew) != 1 || onlyNew[0] != 1 {
		t.Fatalf("second duplicate should be reported as new, got %v", onlyNew)
	}
}

func TestElementKey_PrimitiveStringifies(t *testing.T) {
	k := elementKey("foo", ArrayMatcher{})
	if k != "foo" {
		t.Errorf("primitive key = %q", k)
	}
}

func TestIsOrderedHash(t *testing.T) {
	if !IsOrderedHash("EditorColumnHash") {
		t.Error("EditorColumnHash should be ordered")
	}
	if IsOrderedHash("Anything") {
		t.Error("unknown name should not be ordered")
	}
}

