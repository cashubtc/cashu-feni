package feni

import (
	"fmt"
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	RootCmd.AddCommand(payCommand)

}

var payCommand = &cobra.Command{
	Use:    "pay",
	Short:  "Pay lightning invoice",
	Long:   `Pay a lightning invoice using cashu coins.`,
	PreRun: PreRunFeni,
	Run:    pay,
}

func pay(cmd *cobra.Command, args []string) {
	amount, err := strconv.Atoi(args[0])
	if err != nil {
		panic(err)
	}
	fmt.Println(WalletClient.GetMint(int64(amount)))
}
