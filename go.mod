module github.com/kramerul/shalm

go 1.13

require (
	github.com/Masterminds/sprig/v3 v3.0.0
	github.com/blang/semver v3.5.1+incompatible
	github.com/containerd/containerd v1.3.0
	github.com/deislabs/oras v0.8.0
	github.com/docker/distribution v2.6.0-rc.1.0.20180327202408-83389a148052+incompatible // indirect
	github.com/go-critic/go-critic v0.3.4 // indirect
	github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6 // indirect
	github.com/golangci/go-tools v0.0.0-20190124090046-35a9f45a5db0 // indirect
	github.com/golangci/gofmt v0.0.0-20190930125516-244bba706f1a // indirect
	github.com/golangci/gosec v0.0.0-20180901114220-8afd9cbb6cfb // indirect
	github.com/golangci/lint-1 v0.0.0-20181222135242-d2cdd8c08219 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.2
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.1
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.8.1
	github.com/qri-io/starlib v0.4.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stevvooe/resumable v0.0.0-20180830230917-22b14a53ba50 // indirect
	go.starlark.net v0.0.0-20191021185836-28350e608555
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20191010194322-b09406accb47 // indirect
	gopkg.in/yaml.v2 v2.2.4
	mvdan.cc/unparam v0.0.0-20191111180625-960b1ec0f2c2 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect

)

replace github.com/go-critic/go-critic => github.com/go-critic/go-critic v0.3.4

replace github.com/golangci/errcheck => github.com/golangci/errcheck v0.0.0-20181223084120-ef45e06d44b6

replace github.com/golangci/go-tools => github.com/golangci/go-tools v0.0.0-20190124090046-35a9f45a5db0

replace github.com/golangci/gofmt => github.com/golangci/gofmt v0.0.0-20190930125516-244bba706f1a

replace github.com/golangci/gosec => github.com/golangci/gosec v0.0.0-20180901114220-8afd9cbb6cfb

replace github.com/golangci/lint-1 => github.com/golangci/lint-1 v0.0.0-20181222135242-d2cdd8c08219

replace mvdan.cc/unparam => mvdan.cc/unparam v0.0.0-20191111180625-960b1ec0f2c2
