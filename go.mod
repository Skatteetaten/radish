module github.com/skatteetaten/radish

require (
	github.com/drone/envsubst v1.0.3
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.5
	github.com/pkg/errors v0.9.1
	github.com/plaid/go-envvar v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace (
	github.com/coreos/etcd => github.com/coreos/etcd v3.3.26+incompatible
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/miekg/dns => github.com/miekg/dns v1.1.43
	golang.org/x/text => golang.org/x/text v0.3.7
)

go 1.17
