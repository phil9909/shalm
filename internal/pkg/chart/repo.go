package chart

// Repo -
type Repo interface {
	// Directory -
	Directory(name string) (string, error)
}

// LocalRepo -
type LocalRepo struct {
	BaseDir string
}

// Directory -
func (r *LocalRepo) Directory(name string) (string, error) {
	return r.BaseDir + "/" + name + "/", nil
}
