package feni

import (
	"fmt"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/gohumble/cashu-feni/mint"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"time"
)

var invoiceCommand = &cobra.Command{
	Use:    "invoice",
	Short:  "Creates a new invoice, if lightning is enabled",
	Long:   ``,
	PreRun: PreRunFeni,
	Run:    mintCmd,
}
var hash string

func init() {
	invoiceCommand.PersistentFlags().StringVarP(&hash, "hash", "", "", "the hash of the mint you want to claim")
	RootCmd.AddCommand(invoiceCommand)
}
func mintCmd(cmd *cobra.Command, args []string) {
	amount, err := strconv.Atoi(args[0])
	if err != nil {
		panic(err)
	}
	splitAmount := mint.AmountSplit(uint64(amount))
	if amount > 0 {
		if !Config.Lightning {
			if err := storeProofs(Wallet.mint(splitAmount, hash)); err != nil {
				log.Error(err)
			}
			return
		}
		if hash == "" {
			invoice, err := WalletClient.GetMint(uint64(amount))
			if err != nil {
				panic(err)
			}
			fmt.Printf("Pay invoice to mint %d sat:\n", amount)
			fmt.Printf("Invoice: %s\n", invoice.Pr)
			fmt.Printf("Execute this command if you abort the check:\nfeni invoice {amount} --hash %s\n", invoice.Hash)
			fmt.Printf("Checking invoice ...")
			for {
				time.Sleep(time.Second * 3)
				proofs := Wallet.mint(splitAmount, invoice.Hash)
				if len(proofs) == 0 {
					fmt.Print(".")
					continue
				}
				// storeProofs
				err := storeProofs(proofs)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("Invoice paid.")
				return
			}
		} else {

		}
	}
}

func invalidate(proof cashu.Proof) error {
	err := storage.DeleteProof(proof)
	if err != nil {
		return err
	}
	return storage.StoreUsedProofs(
		cashu.ProofsUsed{
			Secret:   proof.Secret,
			Amount:   proof.Amount,
			C:        proof.C,
			TimeUsed: time.Now(),
		},
	)
}
func storeProofs(proofs []cashu.Proof) error {
	for _, proof := range proofs {
		Wallet.proofs = append(Wallet.proofs, proof)
		err := storage.StoreProof(proof)
		if err != nil {
			return err
		}
	}
	return nil
}
