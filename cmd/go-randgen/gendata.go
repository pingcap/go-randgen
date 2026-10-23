package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/pingcap/go-randgen/compare"
	"github.com/spf13/cobra"
)

var gendataDsns []string
var gendataDsnsFile string

func newGenDataCmd() *cobra.Command {
	gendataCmd := &cobra.Command{
		Use:   "gendata",
		Short: "generate data in specified db by zz",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			gendataDsns, err = resolveCredentialList(credentialInput{
				flagValue: strings.Join(gendataDsns, ","),
				flagSet:   cmd.Flags().Changed("dsns"),
				filePath:  gendataDsnsFile,
				envName:   "GO_RANDGEN_DSNS",
				label:     "dsns",
			})
			if err != nil {
				if err.Error() == "a dsns is required" {
					return fmt.Errorf("At least one dsn are required")
				}
				return err
			}

			return nil
		},
		Run: gendataAction,
	}

	gendataCmd.Flags().StringSliceVar(&gendataDsns, "dsns", nil, "database DSNs (deprecated: use --dsns-file or GO_RANDGEN_DSNS)")
	gendataCmd.Flags().StringVar(&gendataDsnsFile, "dsns-file", "", "path to an owner-only file containing one database DSN per line")
	_ = gendataCmd.Flags().MarkDeprecated("dsns", "use --dsns-file or GO_RANDGEN_DSNS")

	return gendataCmd
}

func gendataAction(cmd *cobra.Command, args []string) {
	ddls, _ := getDdls()

	targetDbs := make([]*sql.DB, 0, len(gendataDsns))
	for _, dsn := range gendataDsns {
		// fixed: change "mysql" -> dbms
		targetDb, err := compare.OpenDBWithRetry(dbms, dsn)
		if err != nil {
			log.Fatalf("connect dsn1 %s error %v\n", compare.RedactDSN(dsn), err)
		}
		targetDbs = append(targetDbs, targetDb)
	}

	errSql, err := compare.ExecSqlsInDbs(ddls, targetDbs...)
	if err != nil {
		log.Printf("Fatal Error: data prepare ddl exec error %v\n", err)
		log.Fatalln(errSql)
	}

	safeDSNs := make([]string, 0, len(gendataDsns))
	for _, dsn := range gendataDsns {
		safeDSNs = append(safeDSNs, compare.RedactDSN(dsn))
	}
	log.Printf("generate data in:\n %v ok\n", safeDSNs)
}
