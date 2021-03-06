package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/pingcap/go-randgen/compare"
	"github.com/spf13/cobra"
	"log"
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
		targetDb, err := compare.OpenDBWithRetry("mysql", dsn)
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
	if len(targetDbs) == 2 {
		var tableNames []string
		tiflashDb := targetDbs[1]
		rows, err := tiflashDb.Query("show tables")
		if err != nil {
			log.Fatalln(err)
		}
		for rows.Next() {
			if rows.Err() != nil {
				log.Fatalln(err)
			}
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				log.Fatalln(err)
			}
			tableNames = append(tableNames, tableName)
		}
		for _, tn := range tableNames {
			_, err := tiflashDb.Exec(fmt.Sprintf("ALTER TABLE %s SET TIFLASH REPLICA 1", tn))
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	log.Printf("generate data in:\n %v ok\n", gendataDsns)
}
