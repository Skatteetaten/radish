module github.com/skatteetaten/radish

go 1.18

// direct dependencies:
require (
	github.com/drone/envsubst v1.0.3
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/magiconair/properties v1.8.6
	github.com/pkg/errors v0.9.1
	github.com/plaid/go-envvar v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.7.1
)

// indirect dependencies:
require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20220513210249-45d2b4557a2a // indirect
	gopkg.in/yaml.v3 v3.0.0-20220512140231-539c8e751b99 // indirect
)

replace (
	github.com/coreos/etcd => github.com/coreos/etcd v3.5.4+incompatible
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
	github.com/miekg/dns => github.com/miekg/dns v1.1.49
	golang.org/x/text => golang.org/x/text v0.3.7
)
