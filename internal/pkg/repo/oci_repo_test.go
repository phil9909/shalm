package repo

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	root       = path.Join(filepath.Dir(b), "..", "..", "..")
)

var _ = Describe("OCIRepo", func() {

	Context("push chart", func() {
		It("pushes chart correct", func() {
			repo := OciRepo{BaseDir: path.Join(root, "example")}
			err := repo.Push("uaa")
			Expect(err).ToNot(HaveOccurred())
		})
		It("oras", func() {
			Expect(test()).NotTo(HaveOccurred())
		})

	})
})

func test() error {
	ref := "gcr.io/peripli/oras:test"
	fileName := "hello.txt"
	fileContent := []byte("Hello World!\n")
	customMediaType := "my.custom.media.type"

	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(
			docker.WithAuthorizer(
				docker.NewDockerAuthorizer(docker.WithAuthCreds(func(repository string) (s string, s2 string, e error) {
					return "_json_key", os.Getenv("GCR_ADMIN_CREDENTIALS"), nil
				})))),
	})

	// Push file(s) w custom mediatype to registry
	memoryStore := content.NewMemoryStore()
	desc := memoryStore.Add(fileName, customMediaType, fileContent)
	pushContents := []ocispec.Descriptor{desc}
	fmt.Printf("Pushing %s to %s...\n", fileName, ref)
	desc, err := oras.Push(ctx, resolver, ref, memoryStore, pushContents)
	if err != nil {
		return err
	}

	// Pull file(s) from registry and save to disk
	fmt.Printf("Pulling from %s and saving to %s...\n", ref, fileName)
	fileStore := content.NewFileStore("")
	defer fileStore.Close()
	allowedMediaTypes := []string{customMediaType}
	desc, _, err = oras.Pull(ctx, resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return err
	}
	fmt.Printf("Pulled from %s with digest %s\n", ref, desc.Digest)
	fmt.Printf("Try running 'cat %s'\n", fileName)
	return nil
}
