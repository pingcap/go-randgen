package main

import (
	"errors"
	"github.com/pingcap/go-randgen/compare"
	"github.com/spf13/cobra"
	"log"
)


var gendataDsn string

func newGenDataCmd() *cobra.Command {
	gendataCmd := &cobra.Command{
		Use:   "gendata",
		Short: "generate data in specified db by zz",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if gendataDsn == "" {
				return errors.New("dsn are required")
			}

			return nil
		},
		Run: gendataAction,
	}

	gendataCmd.Flags().StringVar(&gendataDsn, "dsn", "", "specified db to generate data by zz")

	return gendataCmd
}


func gendataAction(cmd *cobra.Command, args []string) {
	ddls, _ := getDdls()

	targetDb, err := compare.OpenDBWithRetry("mysql", gendataDsn)
	if err != nil {
		log.Fatalf("connect dsn1 %s error %v\n", gendataDsn, err)
	}

	errSql, err := compare.ExecSqlsInDbs(ddls, targetDb)
	if err != nil {
		log.Printf("Fatal Error: data prepare ddl exec error %v\n", err)
		log.Fatalln(errSql)
	}

	log.Printf("generate data in %s ok\n", gendataDsn)
}
