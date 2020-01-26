package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
	"strings"
)

/*
   Creation Time: 2019 - Oct - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func init() {
	RootCmd.AddCommand(DevCmd)
	DevCmd.AddCommand(UnsubscribeCmd)
}

var DevCmd = &cobra.Command{
	Use: "Dev",
}

var UnsubscribeCmd = &cobra.Command{
	Use: "Unsubscribe",
	Run: func(cmd *cobra.Command, args []string) {
		v := url.Values{}
		v.Set("phone", cmd.Flag(FlagPhone).Value.String())
		_, err := sendHttp(http.MethodPost, "dev/unsubscribe", ContentTypeForm,
			strings.NewReader(v.Encode()),
			true,
		)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}
