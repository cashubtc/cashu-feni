package feni

import (
	"fmt"
	"github.com/cashubtc/cashu-feni/cashu"
	"github.com/cashubtc/cashu-feni/db"
	"github.com/cashubtc/cashu-feni/lightning"
	"github.com/cashubtc/cashu-feni/mint"
	"github.com/samber/lo"
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
			var invoice lightning.Invoicer
			invoice, err = Wallet.Client.GetMint(int64(amount))
			if err != nil {
				panic(err)
			}
			err = storage.StoreLightningInvoice(invoice)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Pay invoice to mint %d sat:\n", amount)
			fmt.Printf("Invoice: %s\n", invoice.GetPaymentRequest())
			fmt.Printf("Execute this command if you abort the check:\nfeni invoice {amount} --hash %s\n", invoice.GetHash())
			fmt.Printf("Checking invoice ...")
			for {
				time.Sleep(time.Second * 3)
				proofs := Wallet.mint(splitAmount, invoice.GetHash())
				if len(proofs) == 0 {
					fmt.Print(".")
					continue
				}
				// storeProofs
				err = storeProofs(proofs)
				if err != nil {
					log.Error(err.Error())
				}
				fmt.Println("Invoice paid.")
				err = storage.UpdateLightningInvoice(invoice.GetHash(), db.UpdateInvoicePaid(true))
				if err != nil {
					log.Fatal(err)
				}
				return
			}
		} else {
			Wallet.Mint(uint64(amount), hash)
		}
	}
}

func invalidate(proofs []cashu.Proof) error {
	resp, err := Wallet.Client.Check(cashu.CheckSpendableRequest{Proofs: proofs})
	if err != nil {
		return err
	}
	invalidatedProofs := make([]cashu.Proof, 0)
	for i, spendable := range resp.Spendable {
		if !spendable {
			invalidatedProofs = append(invalidatedProofs, proofs[i])
			err = invalidateProof(proofs[i])
			if err != nil {
				return err
			}
		}
	}
	invalidatedSecrets := make([]string, 0)
	for _, proof := range invalidatedProofs {
		invalidatedSecrets = append(invalidatedSecrets, proof.Secret)
	}
	Wallet.proofs = lo.Filter[cashu.Proof](Wallet.proofs, func(p cashu.Proof, i int) bool {
		_, found := lo.Find[string](invalidatedSecrets, func(secret string) bool {
			return secret == p.Secret
		})
		return !found
	})
	return nil
}
func invalidateProof(proof cashu.Proof) error {
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
