// Package types holds value types shared between ent schemas and the
// generated code: entc requires custom JSON field types to live outside
// the schema package itself, because the generated ent code importing the
// type back would close an import cycle (schema → ent → schema).
package types

// GameEntry is the persisted shape of one preference entry inside the
// profile aggregate: a catalog reference (game_id) or a user-typed name
// awaiting admin classification (custom_name) — exactly one of the pair is
// set. Aggregate-contained value: replace-set semantics, no child identity,
// so the list rides the row instead of owning one.
type GameEntry struct {
	GameID     string `json:"game_id,omitempty"`
	CustomName string `json:"custom_name,omitempty"`
}
