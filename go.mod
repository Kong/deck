module github.com/hbagdi/deck

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/fatih/color v1.7.0
	github.com/hashicorp/go-memdb v1.0.4
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hbagdi/go-kong v0.9.0
	github.com/imdario/mergo v0.3.7
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mitchellh/go-homedir v1.0.0
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/spf13/viper v1.2.1
	github.com/stretchr/testify v1.4.0
	github.com/yudai/gojsondiff v1.0.0
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 // indirect
	golang.org/x/sys v0.0.0-20190405154228-4b34438f7a67 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

go 1.13

replace github.com/hashicorp/go-memdb => github.com/hbagdi/go-memdb v0.0.0-20190920041452-92e457f524d8
