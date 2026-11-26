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
var graphTokenFile string
var graphToken string

func newListenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "debug subcommand for /graph restful interface",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if yyPath == "" {
				return errors.New("yy are required")
			}
			var err error
			graphToken, err = resolveCredential(credentialInput{
				filePath: graphTokenFile,
				envName:  "GO_RANDGEN_GRAPH_TOKEN",
				label:    "graph access token",
			})
			return err
		},
		Run: listenAction,
	}

	listenCmd.Flags().IntVar(&port, "port", 43000, "the port to listen")
	listenCmd.Flags().StringVar(&graphTokenFile, "token-file", "", "owner-only file containing the graph access token")

	return listenCmd
}

func listenAction(cmd *cobra.Command, args []string) {
	handler, err := view.Graph(loadYy(), graphToken)
	if err != nil {
		log.Fatalf("Fatal Error: %v\n", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/graph", handler)
	address := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("listen on %s\n", address)
	log.Fatalln(http.ListenAndServe(address, mux))
}
