module github.com/phouse512/piper-tools

go 1.15

// optional when using local go-coda version
// replace github.com/phouse512/go-coda => /Users/philiphouse/os/go-coda

require github.com/urfave/cli/v2 v2.0.0

require (
	github.com/olekukonko/tablewriter v0.0.4
	github.com/phouse512/go-coda v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/avast/retry-go v3.0.0+incompatible
)
