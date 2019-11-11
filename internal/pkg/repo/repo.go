package repo

// Repo -
type Repo interface {
	// Directory -
	Directory(name string) (string, error)
}
