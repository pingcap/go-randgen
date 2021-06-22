package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/pingcap/go-randgen/view"
	"github.com/spf13/cobra"
)

var port int

func newListenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "debug subcommand for /graph restful interface",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if yyPath == "" {
				return errors.New("yy are required")
			}
			return nil
		},
		Run: listenAction,
	}

	listenCmd.Flags().IntVar(&port, "port", 43000, "the port to listen")

	return listenCmd
}

func listenAction(cmd *cobra.Command, args []string) {
	handler, err := view.Graph(loadYy())
	if err != nil {
		log.Fatalf("Fatal Error: %v\n", err)
	}
	http.HandleFunc("/graph", handler)
	log.Printf("listen on :%d\n", port)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
