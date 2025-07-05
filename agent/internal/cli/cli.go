package cli

type Start struct {
	Interval   string `help:"Interval between runs" default:"5s"`
	Identifier string `help:"Agent identifier" default:"unknown"`
}

type CLI struct {
	Start Start `cmd:"" help:"Start agent"`
}
