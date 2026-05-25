package sitediff

import "strings"

// IgnoreFilter decides whether a given (path, leafKey) should be excluded
// from the diff. Two flavours of rule are supported:
//
//   - Leaf names (e.g. "SiteId"): match the leaf key at any depth.
//   - JSONPath globs (start with "/"): match a literal path. "*" matches a
//     single segment, "**" matches zero or more segments.
//
// A path is segmented purely by "/". Array elements are addressed by their
// numeric index (0, 1, ...) regardless of semantic key.
type IgnoreFilter struct {
	leafKeys map[string]struct{}
	pathPats []ignorePat
}

type ignorePat struct {
	segments []string // raw segments, may contain "*" or "**"
}

// NewIgnoreFilter assembles a filter from Options.
//
// DefaultIgnoreKeys are pre-loaded unless Options.NoDefaultIgnore is set.
// The order of evaluation is leaf-name first, then path-glob; either match
// is enough to ignore.
func NewIgnoreFilter(opts Options) *IgnoreFilter {
	f := &IgnoreFilter{leafKeys: make(map[string]struct{})}

	if !opts.NoDefaultIgnore {
		for _, k := range DefaultIgnoreKeys {
			f.leafKeys[k] = struct{}{}
		}
	}
	for _, k := range opts.IgnoreKeys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		f.leafKeys[k] = struct{}{}
	}
	for _, p := range opts.IgnorePaths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// JSONPath-style. We expect a leading "/", but accept either form.
		p = strings.TrimPrefix(p, "/")
		f.pathPats = append(f.pathPats, ignorePat{segments: strings.Split(p, "/")})
	}
	return f
}

// ShouldIgnore returns true when the location should be excluded from the diff.
// path is "/"-joined; leafKey is the trailing segment.
func (f *IgnoreFilter) ShouldIgnore(path, leafKey string) bool {
	if f == nil {
		return false
	}
	if leafKey != "" {
		if _, ok := f.leafKeys[leafKey]; ok {
			return true
		}
	}
	if len(f.pathPats) == 0 {
		return false
	}
	segs := splitPath(path)
	for _, pat := range f.pathPats {
		if matchSegments(pat.segments, segs) {
			return true
		}
	}
	return false
}

// splitPath turns "/Sites/0/SiteSettings/Columns/2/LabelText" into the
// segment list. Leading slash is tolerated.
func splitPath(p string) []string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// matchSegments performs glob matching with two wildcards:
//
//	"*"  : exactly one segment, any value
//	"**" : zero or more segments
//
// The implementation is a small recursive descent — patterns are short
// (a handful of segments) so this is plenty fast. Returning a separate
// helper keeps ShouldIgnore readable.
func matchSegments(pat, in []string) bool {
	for len(pat) > 0 {
		head := pat[0]
		if head == "**" {
			// Try matching ** against 0..len(in) leading segments of `in`.
			rest := pat[1:]
			if len(rest) == 0 {
				return true // ** at tail consumes everything
			}
			for i := 0; i <= len(in); i++ {
				if matchSegments(rest, in[i:]) {
					return true
				}
			}
			return false
		}
		if len(in) == 0 {
			return false
		}
		if head != "*" && head != in[0] {
			return false
		}
		pat = pat[1:]
		in = in[1:]
	}
	return len(in) == 0
}
