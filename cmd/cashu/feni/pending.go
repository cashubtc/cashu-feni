package feni

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pendingCommand = &cobra.Command{
	Use:    "pending",
	Short:  "Show pending tokens",
	Long:   ``,
	PreRun: PreRunFeni,
	Run:    pendingCmd,
}

func init() {
	RootCmd.AddCommand(pendingCommand)
}
func pendingCmd(cmd *cobra.Command, args []string) {
	reserved, err := storage.GetReservedProofs()
	if err != nil {
		log.Fatal(err)
	}
	if len(reserved) > 0 {
		fmt.Println("--------------------------")
	}
	for i, proof := range reserved {
		fmt.Printf("#%d Amount: %d sat Time: %s, ID: %s\n", i, proof.Amount, proof.TimeReserved, proof.SendId.String())
	}
}
