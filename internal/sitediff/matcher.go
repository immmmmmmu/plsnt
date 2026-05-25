package sitediff

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// ArrayMatcher describes how to align two array snapshots taken at different
// points in time. Each strategy is tried in order; the first one that
// produces a non-empty key for an element wins.
type ArrayMatcher struct {
	// PrimaryKeys are leaf field names. A composite key joins multiple
	// fields with '\x1f' (unit separator). If a field is missing the
	// strategy is skipped.
	PrimaryKeys [][]string
	// FallbackHash, when true, uses the first 80 chars of the "Body" field
	// (or any single string field on the element) as a stability key when
	// no PrimaryKeys produce a value.
	FallbackHash bool
	// Ordered, when true, suppresses semantic matching entirely: elements
	// are paired by index. Use for fields like EditorColumnHash.* whose
	// position carries meaning.
	Ordered bool
}

// DefaultMatchers maps a leaf array name to its matching strategy.
//
// The keys here are the *leaf* names of the array (e.g. "Columns"), not full
// paths. Pleasanter reuses array names consistently across the SitePackage
// schema, so a flat lookup is correct.
//
// Composite keys: the inner slice is tried first as a tuple; if any
// component is missing we fall through to the next slice.
var DefaultMatchers = map[string]ArrayMatcher{
	"Sites":          {PrimaryKeys: [][]string{{"SiteId"}, {"Title", "ParentId"}, {"Title"}}},
	"Columns":        {PrimaryKeys: [][]string{{"ColumnName"}}},
	"Scripts":        {PrimaryKeys: [][]string{{"Title"}, {"Id"}}, FallbackHash: true},
	"Styles":         {PrimaryKeys: [][]string{{"Title"}, {"Id"}}, FallbackHash: true},
	"ServerScripts":  {PrimaryKeys: [][]string{{"Title"}, {"Id"}}, FallbackHash: true},
	"Views":          {PrimaryKeys: [][]string{{"Name"}, {"Id"}}},
	"Processes":      {PrimaryKeys: [][]string{{"Name"}, {"Id"}}},
	"StatusControls": {PrimaryKeys: [][]string{{"Status"}, {"Id"}}},
	"Notifications":  {PrimaryKeys: [][]string{{"Type", "Address"}, {"Id"}}},
	"Reminders":      {PrimaryKeys: [][]string{{"Subject"}, {"Id"}}},
	"Links":          {PrimaryKeys: [][]string{{"ColumnName", "SiteId"}, {"ColumnName"}}},
	"Convertors":     {PrimaryKeys: [][]string{{"SiteId"}, {"SiteTitle"}}},
	"Permissions":    {PrimaryKeys: [][]string{{"Name", "PermissionType"}, {"Name"}}},
	// Ordered arrays: position carries meaning, semantic match would corrupt diffs.
	"GridColumns": {Ordered: true},
}

// orderedNamedArrays additionally lists arrays that are *children of named
// hash maps* (EditorColumnHash, FilterColumnHash, ...). The key here is the
// containing hash name; everything inside is treated as a list of strings
// where order matters.
var orderedNamedArrays = map[string]bool{
	"EditorColumnHash": true,
	"FilterColumnHash": true,
	"TabSettings":      true,
}

// ArrayMatch binds an old-side index to a new-side index plus the key value
// that produced the match (for diagnostics and stable ordering).
type ArrayMatch struct {
	OldIndex int
	NewIndex int
	Key      string
}

// MatchArrays aligns old[] and new[] by semantic key. arrayName drives the
// strategy lookup; an unknown name (or an Ordered matcher) falls back to
// index matching.
//
// Each PrimaryKeys group is tried as a separate pass over the *remaining*
// pool of unmatched elements. This means when one side renames or drops a
// preferred field, we can still match via a later fallback (e.g. Title
// missing → match by Id). Within a pass, duplicate keys get a first-wins
// pairing; remaining duplicates surface as Added/Removed for the next
// pass to resolve.
//
// Returns:
//   - matched : pairs of indices that share a key, sorted by OldIndex
//   - onlyOld : indices present only in old
//   - onlyNew : indices present only in new
func MatchArrays(arrayName string, old, newArr []any) (matched []ArrayMatch, onlyOld, onlyNew []int) {
	m, ok := DefaultMatchers[arrayName]
	if !ok || m.Ordered {
		return matchByIndex(old, newArr)
	}

	oldPool := makeRange(len(old))
	newPool := makeRange(len(newArr))

	tryStrategy := func(keyFn func(any) string) {
		newKeyIdx := make(map[string][]int)
		for _, ni := range newPool {
			k := keyFn(newArr[ni])
			if k == "" {
				continue
			}
			newKeyIdx[k] = append(newKeyIdx[k], ni)
		}
		consumed := make(map[int]bool)
		var oldNext []int
		for _, oi := range oldPool {
			k := keyFn(old[oi])
			if k == "" {
				oldNext = append(oldNext, oi)
				continue
			}
			idxs := newKeyIdx[k]
			if len(idxs) == 0 {
				oldNext = append(oldNext, oi)
				continue
			}
			ni := idxs[0]
			newKeyIdx[k] = idxs[1:]
			consumed[ni] = true
			matched = append(matched, ArrayMatch{OldIndex: oi, NewIndex: ni, Key: k})
		}
		oldPool = oldNext
		var newNext []int
		for _, ni := range newPool {
			if !consumed[ni] {
				newNext = append(newNext, ni)
			}
		}
		newPool = newNext
	}

	for _, group := range m.PrimaryKeys {
		group := group // capture
		tryStrategy(func(v any) string {
			obj, ok := v.(map[string]any)
			if !ok {
				return ""
			}
			k, _ := compositeKey(obj, group)
			return k
		})
	}
	if m.FallbackHash {
		tryStrategy(func(v any) string {
			obj, ok := v.(map[string]any)
			if !ok {
				return ""
			}
			h := bodyHash(obj)
			if h == "" {
				return ""
			}
			return "hash:" + h
		})
	}

	sort.Slice(matched, func(i, j int) bool { return matched[i].OldIndex < matched[j].OldIndex })
	return matched, oldPool, newPool
}

func makeRange(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// matchByIndex pairs elements by physical position. Used for Ordered
// arrays and for any array whose name is not in DefaultMatchers.
func matchByIndex(old, newArr []any) (matched []ArrayMatch, onlyOld, onlyNew []int) {
	n := len(old)
	if len(newArr) < n {
		n = len(newArr)
	}
	for i := 0; i < n; i++ {
		matched = append(matched, ArrayMatch{OldIndex: i, NewIndex: i, Key: fmt.Sprintf("[%d]", i)})
	}
	for i := n; i < len(old); i++ {
		onlyOld = append(onlyOld, i)
	}
	for i := n; i < len(newArr); i++ {
		onlyNew = append(onlyNew, i)
	}
	return matched, onlyOld, onlyNew
}

// elementKey reduces an array element to a stable string key per the
// matcher's strategy, evaluating each PrimaryKeys group in order. Returns
// "" when no strategy produces a value. Exported for tests; the main
// MatchArrays uses a per-pass strategy walk instead.
func elementKey(v any, m ArrayMatcher) string {
	obj, ok := v.(map[string]any)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	for _, group := range m.PrimaryKeys {
		if k, ok := compositeKey(obj, group); ok {
			return k
		}
	}
	if m.FallbackHash {
		if s := bodyHash(obj); s != "" {
			return "hash:" + s
		}
	}
	return ""
}

// compositeKey returns a "v1\x1fv2..." joined value when every requested
// field exists and is non-zero. The unit separator avoids collisions with
// values containing "/" or "=".
func compositeKey(obj map[string]any, fields []string) (string, bool) {
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		raw, ok := obj[f]
		if !ok || raw == nil {
			return "", false
		}
		s := stringifyScalar(raw)
		if s == "" {
			return "", false
		}
		parts = append(parts, f+"="+s)
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, "\x1f"), true
}

// bodyHash hashes the first 80 characters of the first usable string field
// (Body / Script / Content) on the object, so duplicates with the same
// nominal Title still get a stable identity.
func bodyHash(obj map[string]any) string {
	for _, f := range []string{"Body", "Script", "Content"} {
		if v, ok := obj[f].(string); ok && v != "" {
			head := v
			if len(head) > 80 {
				head = head[:80]
			}
			sum := sha256.Sum256([]byte(head))
			return hex.EncodeToString(sum[:8]) // 16 hex chars is plenty
		}
	}
	return ""
}

// stringifyScalar renders a JSON scalar (json.Number, string, bool, etc.)
// as the canonical text form used in keys.
func stringifyScalar(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// IsOrderedHash reports whether values *inside* a named hash map should be
// compared by position (not semantically). Used by the recursive walker
// when stepping into EditorColumnHash and friends.
func IsOrderedHash(name string) bool {
	return orderedNamedArrays[name]
}
