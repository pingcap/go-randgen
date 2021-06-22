package main

import (
	"fmt"
	"log"

	"github.com/pingcap/go-randgen/resource"
	"github.com/spf13/cobra"
)

func newZzCmd() *cobra.Command {
	buildInZzCmd := &cobra.Command{
		Use:   "zz",
		Short: "print built-in zz",
		Run: func(cmd *cobra.Command, args []string) {
			bytes, err := resource.Asset("resource/default.zz.lua")
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(string(bytes))
		},
	}

	return buildInZzCmd
}
