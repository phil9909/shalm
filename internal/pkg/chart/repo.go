package chart

// Repo -
type Repo struct {
	directory string
}

var repo = Repo{directory: "example"}

// Directory -
func (r *Repo) Directory(name string) string {
	return r.directory + "/" + name + "/"
}
