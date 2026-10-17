package security

// The population archetype vocabulary is the compile-time half of the
// population axis (pattern: small vocabulary at compile time, instances as
// data rows). engine = organizational grants semantics; patron =
// identity-is-permission. A third word (service, non-human principals) has
// a reserved slot but no resident until its second consumer exists.
const (
	ArchetypeEngine = "engine"
	ArchetypePatron = "patron"
)

// ValidArchetype reports membership in the compile-time vocabulary — the
// single source every archetype input validates against.
func ValidArchetype(a string) bool {
	return a == ArchetypeEngine || a == ArchetypePatron
}

// DefaultPopulation is the first engine instance and the degenerate mode's
// whole vocabulary: products that never configure populations run everyone
// as this single row, unaware it exists.
const DefaultPopulation = "user"
