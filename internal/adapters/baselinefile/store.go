package baselinefile

// Store reads and writes baseline files under a trust policy. It satisfies
// ports.BaselineStore.
type Store struct {
	trust Trust
	force bool
}

// NewStore returns a store that applies the given trust policy when loading.
func NewStore(trust Trust) Store {
	return Store{trust: trust}
}

// Forcing returns a store whose Save skips the check that an existing file is
// a trusted, valid baseline. It never skips the regular-file and symlink
// checks, so it cannot be used to write through a symlink or over a directory.
func (store Store) Forcing() Store {
	store.force = true
	return store
}
