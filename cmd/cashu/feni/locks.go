package feni

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/gohumble/cashu-feni/cashu"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(locksCommand)

}

const getLocksAnnotationValue = "GetLocks"

var GetLocksDynamic = func(annotationValue string) []prompt.Suggest {
	scripts, err := storage.GetScripts("")
	if err != nil {
		return nil
	}
	suggestions := make([]prompt.Suggest, 0)
	for _, script := range scripts {
		suggestions = append(suggestions, prompt.Suggest{Text: fmt.Sprintf("P2SH:%s", script.Address), Description: fmt.Sprintf("Your P2SH lock for receiving and sending cashu coins")})

	}
	return suggestions
}
var locksCommand = &cobra.Command{
	Use:    "locks",
	Short:  "Generate receiving lock",
	Long:   `Generates a receiving lock for cashu coins.`,
	PreRun: PreRunFeni,
	Annotations: map[string]string{
		DynamicSuggestionsAnnotation: getLocksAnnotationValue,
	},
	Run: locks,
}

func locks(cmd *cobra.Command, args []string) {
	fmt.Println(getP2SHLocks())
}

func getP2SHLocks() []cashu.P2SHScript {
	scripts, err := storage.GetScripts("")
	if err != nil {
		return nil
	}
	return scripts
}
