package observability

import (
	stderrors "errors"
	"strings"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
)

// errorView is the flattened projection of an error chain for observability
// outlets: identity (code/reason/message) from the outermost kratos layer,
// the internal classifications carried by kratos layers beneath it, and the
// factual root at the bottom. Cause never crosses the wire boundary, so the
// client sees the fuzzy reason while collection sees the specific one.
type errorView struct {
	code     int
	reason   string
	message  string
	internal []string
	root     string
}

// viewError flattens err into an errorView. Identity comes from the first
// *errors.Error on the chain; kratos layers beneath it contribute their
// reasons as internal classifications; the root is the message of the deepest
// non-kratos node. Raw Error() rendering is never used — kratos v3 inlines
// the whole chain recursively, one "error: code = ... cause = ..." prefix per
// layer, which is what this projection exists to replace. A kratos-only chain
// with no cause — a domain rule violation — has no root: the reason alone is
// the full story.
func viewError(err error) errorView {
	var v errorView

	var outer *kratoserrors.Error
	if stderrors.As(err, &outer) {
		v.code = int(outer.Code)
		v.reason = outer.Reason
		v.message = outer.Message
	}

	var deepest error
	for cur := err; cur != nil; cur = stderrors.Unwrap(cur) {
		// A rendered shell exists only to re-render Error() for downstream
		// observers; its flat text must not be re-parsed as an authored
		// classification layer, so the walk steps straight through it.
		//nolint:errorlint // per-level identity is the point: this walk dissects the chain one layer at a time
		if _, ok := cur.(*renderedError); ok {
			continue
		}
		//nolint:errorlint // same deliberate per-level match; errors.As would re-walk the chain below cur
		if inner, ok := cur.(*kratoserrors.Error); ok {
			if inner != outer {
				v.internal = append(v.internal, inner.Reason)
			}
			continue
		}
		deepest = cur
	}
	if deepest != nil {
		v.root = deepest.Error()
	}
	return v
}

// internalPath joins the internal classifications outermost-first, mirroring
// the chain's direction.
func (v errorView) internalPath() string {
	return strings.Join(v.internal, " > ")
}

// flat renders the single-line form used wherever an error's text reaches an
// observer: REASON [INTERNAL > PATH] (client message): root fact.
func (v errorView) flat() string {
	var b strings.Builder
	b.WriteString(v.reason)
	if len(v.internal) > 0 {
		b.WriteString(" [")
		b.WriteString(v.internalPath())
		b.WriteString("]")
	}
	if v.message != "" {
		b.WriteString(" (")
		b.WriteString(v.message)
		b.WriteString(")")
	}
	if v.root != "" {
		if b.Len() > 0 {
			b.WriteString(": ")
		}
		b.WriteString(v.root)
	}
	if b.Len() == 0 {
		return "unclassified error"
	}
	return b.String()
}

// renderedError re-renders an error's Error() as the flattened observability
// view while delegating every identity query — errors.As walks, Unwrap chains,
// the wire translators' FromError — to the original error unchanged. Handlers
// up the chain (the stock tracing middleware's RecordError/SetStatus, the
// logging middleware, sanitize, ErrorToConnect) keep their exact semantics;
// only the human-facing rendering changes, from kratos v3's recursive chain
// dump to one line. Wrapping is idempotent: an already-wrapped chain passes
// through.
type renderedError struct {
	original error
	view     errorView
}

func (e *renderedError) Error() string { return e.view.flat() }
func (e *renderedError) Unwrap() error { return e.original }

// renderFlat wraps err so its Error() renders the flattened view. Idempotent
// on already-wrapped chains; nil stays nil.
func renderFlat(err error) error {
	if err == nil {
		return nil
	}
	var re *renderedError
	if stderrors.As(err, &re) {
		return err
	}
	return &renderedError{original: err, view: viewError(err)}
}
