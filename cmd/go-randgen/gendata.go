package main

import (
	"database/sql"
	"errors"
	"log"

	"github.com/pingcap/go-randgen/compare"
	"github.com/spf13/cobra"
)

var gendataDsns []string

func newGenDataCmd() *cobra.Command {
	gendataCmd := &cobra.Command{
		Use:   "gendata",
		Short: "generate data in specified db by zz",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(gendataDsns) == 0 {
				return errors.New("At least one dsn are required")
			}

			return nil
		},
		Run: gendataAction,
	}

	gendataCmd.Flags().StringSliceVar(&gendataDsns, "dsns", nil, "specified dbs to generate data by zz")

	return gendataCmd
}

func gendataAction(cmd *cobra.Command, args []string) {
	ddls, _ := getDdls()

	targetDbs := make([]*sql.DB, 0, len(gendataDsns))
	for _, dsn := range gendataDsns {
		// fixed: change "mysql" -> dbms
		targetDb, err := compare.OpenDBWithRetry(dbms, dsn)
		if err != nil {
			log.Fatalf("connect dsn1 %s error %v\n", dsn, err)
		}
		targetDbs = append(targetDbs, targetDb)
	}

	errSql, err := compare.ExecSqlsInDbs(ddls, targetDbs...)
	if err != nil {
		log.Printf("Fatal Error: data prepare ddl exec error %v\n", err)
		log.Fatalln(errSql)
	}

	log.Printf("generate data in:\n %v ok\n", gendataDsns)
}
