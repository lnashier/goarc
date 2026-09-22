package log

import "log/slog"

// attrSegment is either an open group (WithGroup) or a bound set of
// attributes (WithAttrs), recorded in the order they were applied so their
// nesting can be reconstructed when a record is finally emitted.
type attrSegment struct {
	group string      // non-empty for a WithGroup segment
	attrs []slog.Attr // non-empty for a WithAttrs segment
}

// buildAttrs replays segments (from WithAttrs/WithGroup) followed by a
// record's own attributes into a single nested map[string]any, one map per
// open group. A group that never ends up with any attributes underneath it
// — from segments or the record — is omitted entirely, per slog.Handler's
// documented elision rule.
func buildAttrs(segments []attrSegment, recordAttrs []slog.Attr) map[string]any {
	if len(segments) == 0 && len(recordAttrs) == 0 {
		return nil
	}

	root := map[string]any{}
	var path []string

	descend := func() map[string]any {
		cur := root
		for _, name := range path {
			next, ok := cur[name].(map[string]any)
			if !ok {
				next = map[string]any{}
				cur[name] = next
			}
			cur = next
		}
		return cur
	}

	for _, seg := range segments {
		if seg.group != "" {
			path = append(path, seg.group)
			continue
		}
		if len(seg.attrs) == 0 {
			continue
		}
		cur := descend()
		for _, a := range seg.attrs {
			addAttr(cur, a)
		}
	}
	if len(recordAttrs) > 0 {
		cur := descend()
		for _, a := range recordAttrs {
			addAttr(cur, a)
		}
	}

	if len(root) == 0 {
		return nil
	}
	return root
}

// addAttr adds a into dst, following slog's rules for empty keys and
// groups: an Attr with an empty key is dropped unless its value is a
// non-empty group, in which case the group's attributes are inlined at the
// current level; an empty group is elided entirely.
func addAttr(dst map[string]any, a slog.Attr) {
	a.Value = a.Value.Resolve()

	if a.Value.Kind() == slog.KindGroup {
		groupAttrs := a.Value.Group()
		if len(groupAttrs) == 0 {
			return
		}
		if a.Key == "" {
			for _, ga := range groupAttrs {
				addAttr(dst, ga)
			}
			return
		}
		sub, ok := dst[a.Key].(map[string]any)
		if !ok {
			sub = map[string]any{}
			dst[a.Key] = sub
		}
		for _, ga := range groupAttrs {
			addAttr(sub, ga)
		}
		return
	}

	if a.Key == "" {
		return
	}
	dst[a.Key] = a.Value.Any()
}
