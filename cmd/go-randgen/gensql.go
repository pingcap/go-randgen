package main

import (
	"errors"
	"log"
	"math"

	"github.com/pingcap/go-randgen/compare"
	"github.com/pingcap/go-randgen/gendata"
	"github.com/spf13/cobra"
)

var gensqlDsn string

func newGensqlCmd() *cobra.Command {
	gensqlCmd := &cobra.Command{
		Use:   "gensql",
		Short: "random generate sqls by yy, parse yy keyword by user specified db, other then zz file",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if gensqlDsn == "" {
				return errors.New("a dsn is required")
			}

			if queries < 0 {
				return errors.New("if not exec, queries num must be positive or zero")
			}

			if maxRecursive <= 0 {
				maxRecursive = math.MaxInt32
			}
			return nil
		},
		Run: gensqlAction,
	}

	gensqlCmd.Flags().StringVar(&gensqlDsn, "dsn", "", "user specified db")

	return gensqlCmd
}

func gensqlAction(cmd *cobra.Command, args []string) {
	db, err := compare.OpenDBWithRetry(dbms, gensqlDsn)
	log.Println("DBMS is:", dbms)

	if err != nil {
		log.Fatalf("connect to dsn %s fail, %v\n", gensqlDsn, err)
	}

	log.Println("Cache database meta info...")
	keyf, err := gendata.ByDb(db, dbms)
	if err != nil {
		log.Fatalf("Fatal Error: %v\n", err)
	}
	log.Println("Cache database meta info ok, start generate sqls by yy")

	randSqls := getRandSqls(keyf)
	dumpRandSqls(randSqls)
}
