package sitediff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// Diff is the package entry point. It compares two parsed SitePackage JSON
// trees (each a top-level map[string]any) and returns the structured diff.
//
// The function is pure: no I/O, no global state. It is safe to call
// concurrently with distinct inputs.
func Diff(old, newPkg map[string]any, opts Options) (*SiteDiff, error) {
	if old == nil || newPkg == nil {
		return nil, fmt.Errorf("sitediff: both old and new must be non-nil")
	}
	filter := NewIgnoreFilter(opts)
	d := &SiteDiff{}

	// HeaderInfo: metadata is dropped via DefaultIgnoreKeys, exporter flags
	// (IncludeXxx) survive and surface as HeaderChanges.
	d.HeaderChanges = diffHeader(old["HeaderInfo"], newPkg["HeaderInfo"], "HeaderInfo", filter)

	// Sites[]: the meaty part.
	added, removed, moved, modified := diffSites(
		toArray(old["Sites"]), toArray(newPkg["Sites"]),
		opts, filter,
	)
	d.Added = added
	d.Removed = removed
	d.Moved = moved
	d.Modified = modified

	// Permissions: opt-in only. Disabled by default because individual user
	// IDs produce huge noisy diffs.
	if opts.IncludePermissions {
		permChanges := diffArray("Permissions",
			toArray(old["Permissions"]), toArray(newPkg["Permissions"]),
			path{"/Permissions", "Permissions"}, filter,
		)
		d.HeaderChanges = append(d.HeaderChanges, permChanges...)
	}

	return d, nil
}

// path is the running breadcrumb during the recursive walk. It carries two
// representations: a slash-separated form for IgnoreFilter (which expects
// JSONPath-style globs) and a display form for FieldChange.Path output.
type path struct {
	slash string // "/Sites/0/SiteSettings/Columns/3"
	disp  string // "Sites[SiteId=12].SiteSettings.Columns[ColumnName=Status]"
}

func (p path) field(name string) path {
	return path{
		slash: joinSlash(p.slash, name),
		disp:  joinDisp(p.disp, name),
	}
}
func (p path) index(i int, semanticKey string) path {
	idx := strconv.Itoa(i)
	disp := "[" + idx + "]"
	if semanticKey != "" {
		disp = "[" + semanticKey + "]"
	}
	return path{
		slash: p.slash + "/" + idx,
		disp:  p.disp + disp,
	}
}
func joinSlash(base, leaf string) string {
	if base == "" {
		return "/" + leaf
	}
	return base + "/" + leaf
}
func joinDisp(base, leaf string) string {
	if base == "" {
		return leaf
	}
	return base + "." + leaf
}

// leafName returns the trailing segment of a slash path (used by the
// IgnoreFilter leaf-name rule).
func leafName(slash string) string {
	if i := strings.LastIndex(slash, "/"); i >= 0 {
		return slash[i+1:]
	}
	return slash
}

// ----------------------------------------------------------------------
// HeaderInfo
// ----------------------------------------------------------------------

func diffHeader(oldRaw, newRaw any, name string, filter *IgnoreFilter) []FieldChange {
	oldM, _ := oldRaw.(map[string]any)
	newM, _ := newRaw.(map[string]any)
	if oldM == nil && newM == nil {
		return nil
	}
	return diffMap(oldM, newM, path{"/" + name, name}, filter)
}

// ----------------------------------------------------------------------
// Sites
// ----------------------------------------------------------------------

func diffSites(old, newArr []any, opts Options, filter *IgnoreFilter) (
	added []SiteRef, removed []SiteRef, moved []SiteMove, modified []SitePatch,
) {
	matched, onlyOld, onlyNew := matchSites(old, newArr, opts.MatchSitesBy)

	for _, oi := range onlyOld {
		if obj, ok := old[oi].(map[string]any); ok {
			removed = append(removed, siteRefFromMap(obj))
		}
	}
	for _, ni := range onlyNew {
		if obj, ok := newArr[ni].(map[string]any); ok {
			added = append(added, siteRefFromMap(obj))
		}
	}

	for _, m := range matched {
		oldSite, _ := old[m.OldIndex].(map[string]any)
		newSite, _ := newArr[m.NewIndex].(map[string]any)
		if oldSite == nil || newSite == nil {
			continue
		}

		// Moved detection happens BEFORE leaf ignore is applied, because
		// ParentId might be on the user's --ignore list (or simply on the
		// default list in a future version). We snapshot the raw values here.
		oldParent := asInt64(oldSite["ParentId"])
		newParent := asInt64(newSite["ParentId"])
		if oldParent != newParent {
			ref := siteRefFromMap(newSite)
			moved = append(moved, SiteMove{SiteRef: ref, OldParentID: oldParent, NewParentID: newParent})
		}

		ref := siteRefFromMap(newSite)
		base := path{
			slash: "/Sites/" + strconv.Itoa(m.NewIndex),
			disp:  fmt.Sprintf("Sites[%s]", m.Key),
		}
		changes := diffMap(oldSite, newSite, base, filter)
		if len(changes) > 0 {
			modified = append(modified, SitePatch{SiteRef: ref, Changes: changes})
		}
	}

	sort.Slice(added, func(i, j int) bool { return added[i].SiteID < added[j].SiteID })
	sort.Slice(removed, func(i, j int) bool { return removed[i].SiteID < removed[j].SiteID })
	sort.Slice(moved, func(i, j int) bool { return moved[i].SiteID < moved[j].SiteID })
	sort.Slice(modified, func(i, j int) bool { return modified[i].SiteID < modified[j].SiteID })
	return
}

// matchSites picks a key strategy for Sites[] depending on opts.MatchSitesBy.
// "auto" (the default) tries SiteId first and falls back to Title+ParentId
// for elements that didn't get matched (handy for cross-environment exports
// where SiteIds were regenerated).
func matchSites(old, newArr []any, mode string) ([]ArrayMatch, []int, []int) {
	switch strings.ToLower(mode) {
	case "title":
		// Title-only first, then Title+ParentId for disambiguation.
		fakeMatcher := ArrayMatcher{PrimaryKeys: [][]string{{"Title", "ParentId"}, {"Title"}}}
		return matchWithMatcher(old, newArr, fakeMatcher)
	case "siteid":
		fakeMatcher := ArrayMatcher{PrimaryKeys: [][]string{{"SiteId"}}}
		return matchWithMatcher(old, newArr, fakeMatcher)
	default: // "auto" or empty
		return MatchArrays("Sites", old, newArr)
	}
}

// matchWithMatcher reuses the cascade logic from MatchArrays with a custom
// matcher (no DefaultMatchers lookup).
func matchWithMatcher(old, newArr []any, m ArrayMatcher) ([]ArrayMatch, []int, []int) {
	// Quick adapter: register a temp name in DefaultMatchers? No — we copy
	// the strategy walk inline to avoid mutating the global table.
	oldPool := makeRange(len(old))
	newPool := makeRange(len(newArr))
	var matched []ArrayMatch

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
		group := group
		tryStrategy(func(v any) string {
			obj, ok := v.(map[string]any)
			if !ok {
				return ""
			}
			k, _ := compositeKey(obj, group)
			return k
		})
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].OldIndex < matched[j].OldIndex })
	return matched, oldPool, newPool
}

func siteRefFromMap(m map[string]any) SiteRef {
	return SiteRef{
		SiteID: asInt64(m["SiteId"]),
		Title:  asString(m["Title"]),
	}
}

// ----------------------------------------------------------------------
// Recursive value comparison
// ----------------------------------------------------------------------

// diffValue dispatches on the runtime type of old/new. Either side may be
// nil (representing an absent key); we distinguish that case from explicit
// json null using the calling diffMap, which checks key presence directly.
func diffValue(oldV, newV any, p path, filter *IgnoreFilter) []FieldChange {
	if filter.ShouldIgnore(p.slash, leafName(p.slash)) {
		return nil
	}

	switch ov := oldV.(type) {
	case map[string]any:
		if nv, ok := newV.(map[string]any); ok {
			return diffMap(ov, nv, p, filter)
		}
	case []any:
		if nv, ok := newV.([]any); ok {
			return diffArray(leafName(p.slash), ov, nv, p, filter)
		}
	}

	// Type mismatch or scalar comparison.
	if equalScalar(oldV, newV) {
		return nil
	}
	return []FieldChange{makeScalarChange(p, oldV, newV)}
}

func diffMap(old, newMap map[string]any, p path, filter *IgnoreFilter) []FieldChange {
	if old == nil && newMap == nil {
		return nil
	}
	if old == nil {
		old = map[string]any{}
	}
	if newMap == nil {
		newMap = map[string]any{}
	}

	// Stable key order: union(old, new), sorted.
	keys := unionKeys(old, newMap)
	var out []FieldChange
	for _, k := range keys {
		oldV, oldOk := old[k]
		newV, newOk := newMap[k]
		child := p.field(k)

		if filter.ShouldIgnore(child.slash, k) {
			continue
		}

		switch {
		case oldOk && !newOk:
			out = append(out, FieldChange{Path: child.disp, Kind: ChangeRemoved, OldValue: oldV})
		case !oldOk && newOk:
			out = append(out, FieldChange{Path: child.disp, Kind: ChangeAdded, NewValue: newV})
		default:
			out = append(out, diffValue(oldV, newV, child, filter)...)
		}
	}
	return out
}

func diffArray(name string, old, newArr []any, p path, filter *IgnoreFilter) []FieldChange {
	matched, onlyOld, onlyNew := MatchArrays(name, old, newArr)
	var out []FieldChange

	// Removed elements first (stable display order: by old index).
	sort.Ints(onlyOld)
	for _, oi := range onlyOld {
		child := p.index(oi, "")
		out = append(out, FieldChange{Path: child.disp, Kind: ChangeRemoved, OldValue: old[oi]})
	}
	for _, m := range matched {
		semantic := semanticDisplay(name, newArr[m.NewIndex], m)
		child := path{
			slash: p.slash + "/" + strconv.Itoa(m.NewIndex),
			disp:  p.disp + "[" + semantic + "]",
		}
		out = append(out, diffValue(old[m.OldIndex], newArr[m.NewIndex], child, filter)...)
	}
	sort.Ints(onlyNew)
	for _, ni := range onlyNew {
		child := p.index(ni, "")
		out = append(out, FieldChange{Path: child.disp, Kind: ChangeAdded, NewValue: newArr[ni]})
	}
	return out
}

// semanticDisplay turns an ArrayMatch into a human-readable selector.
// For Columns it produces "ColumnName=Status"; for indexed fallbacks it
// produces "[i]". The form is used in FieldChange.Path so reviewers can
// jump straight to the right element.
func semanticDisplay(arrayName string, newElem any, m ArrayMatch) string {
	if m.Key != "" && !strings.HasPrefix(m.Key, "[") {
		return m.Key
	}
	return strconv.Itoa(m.NewIndex)
}

// ----------------------------------------------------------------------
// Scalar comparison and unified diff
// ----------------------------------------------------------------------

func equalScalar(a, b any) bool {
	// Treat (nil) and (json null) the same. json.Unmarshal puts json null as nil.
	if a == nil && b == nil {
		return true
	}
	// json.Number equality: compare textual form (preserves precision).
	an, aok := a.(json.Number)
	bn, bok := b.(json.Number)
	if aok && bok {
		return an.String() == bn.String()
	}
	if aok {
		return an.String() == fmt.Sprintf("%v", b)
	}
	if bok {
		return fmt.Sprintf("%v", a) == bn.String()
	}
	return reflect.DeepEqual(a, b)
}

func makeScalarChange(p path, oldV, newV any) FieldChange {
	leaf := leafName(p.slash)

	// Multi-line strings get a unified diff body.
	oldS, oldIsStr := oldV.(string)
	newS, newIsStr := newV.(string)
	if oldIsStr && newIsStr && shouldUnifiedDiff(leaf, oldS, newS) {
		return FieldChange{
			Path:        p.disp,
			Kind:        ChangeTextDiff,
			OldValue:    oldS,
			NewValue:    newS,
			UnifiedDiff: unifiedDiff(oldS, newS),
		}
	}
	return FieldChange{Path: p.disp, Kind: ChangeModified, OldValue: oldV, NewValue: newV}
}

// forceTextDiffFields are leaf names whose value should always be rendered
// as a unified diff regardless of length, because they routinely hold
// formatted multi-line content.
var forceTextDiffFields = map[string]bool{
	"Body":            true,
	"Script":          true,
	"HtmlTitleTop":    true,
	"HtmlTitleSite":   true,
	"HtmlTitleRecord": true,
	"GridGuide":       true,
	"EditorGuide":     true,
	"CalendarGuide":   true,
	"CrosstabGuide":   true,
	"GanttGuide":      true,
	"BurnDownGuide":   true,
	"TimeSeriesGuide": true,
	"AnalyGuide":      true,
	"KambanGuide":     true,
	"ImageLibGuide":   true,
	"ChoicesText":     true,
	"Description":     true,
	"DefaultInput":    true,
}

// forcePlainFields suppresses unified diff for short fields that may
// occasionally contain a stray newline. Unified diff for these is
// noisier than a simple "old → new" line.
var forcePlainFields = map[string]bool{
	"Title":     true,
	"Name":      true,
	"LabelText": true,
	"SiteName":  true,
}

func shouldUnifiedDiff(leafKey, oldS, newS string) bool {
	if forcePlainFields[leafKey] {
		return false
	}
	if forceTextDiffFields[leafKey] {
		return true
	}
	hasNL := strings.Contains(oldS, "\n") || strings.Contains(newS, "\n")
	if !hasNL {
		return false
	}
	return len(oldS) > 60 || len(newS) > 60
}

func unifiedDiff(oldS, newS string) string {
	d := difflib.UnifiedDiff{
		A:        difflib.SplitLines(oldS),
		B:        difflib.SplitLines(newS),
		FromFile: "old",
		ToFile:   "new",
		Context:  3,
	}
	var buf bytes.Buffer
	if err := difflib.WriteUnifiedDiff(&buf, d); err != nil {
		// Failure here is not user-actionable; fall back to a plain marker.
		return fmt.Sprintf("(unified diff failed: %v)", err)
	}
	return buf.String()
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

func toArray(v any) []any {
	a, _ := v.([]any)
	return a
}

func unionKeys(a, b map[string]any) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		seen[k] = struct{}{}
	}
	for k := range b {
		seen[k] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func asInt64(v any) int64 {
	switch x := v.(type) {
	case json.Number:
		n, err := x.Int64()
		if err == nil {
			return n
		}
		// Allow integral floats like "12345.0".
		if f, err2 := x.Float64(); err2 == nil {
			return int64(f)
		}
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
