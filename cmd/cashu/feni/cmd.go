package feni

import (
	"github.com/cashubtc/cashu-feni/wallet"
	"github.com/spf13/cobra"
)

type command func(wallet *wallet.Wallet, params cobraParameter)
type cobraParameter struct {
	cmd  *cobra.Command
	args []string
}
type RootCommand struct {
	cmd    *cobra.Command
	wallet *wallet.Wallet
}

func (r *RootCommand) SetCommand(cmd *cobra.Command) {
	r.cmd = cmd
}
func (r *RootCommand) Command() *cobra.Command {
	return r.cmd
}
func (r *RootCommand) Wallet() *wallet.Wallet {
	return r.wallet
}
func RunCommandWithWallet(root *RootCommand, command command) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		cmd.ParseFlags(args)
		command(root.wallet, cobraParameter{cmd: cmd, args: args})
	}
}
func preRun(w *wallet.Wallet, params cobraParameter) {
	opts := make([]wallet.Option, 0)
	if WalletName != "" {
		opts = append(opts, wallet.WithName(WalletName))
	}
	RootCmd.wallet = wallet.New(opts...)
}
