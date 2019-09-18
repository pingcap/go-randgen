package main

import (
	"errors"
	"github.com/spf13/cobra"
)

var dsn1 string
var dsn2 string

func newExecCmd() *cobra.Command {
	execCmd := &cobra.Command{
		Use:"exec",
		Short:"exec sql in two dsn and compare their result",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if yyPath == "" {
				return errors.New("yy are required")
			}
			if (dsn1 == "" && dsn2 != "") || (dsn1 != "" && dsn2 == "") {
				return errors.New("dsn must have a pair")
			}

			return nil
		},
		Run: execAction,
	}

	execCmd.Flags().StringVar(&dsn1, "dsn1", "", "one of compare dsn")
	execCmd.Flags().StringVar(&dsn2, "dsn2", "", "one of compare dsn")

	return execCmd
}

func execAction(cmd *cobra.Command, args []string)  {
	if dsn1 != "" && dsn2 != "" {
		// compare two dsn
	}
}
