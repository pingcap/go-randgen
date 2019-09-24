package main

import (
	"errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var format bool
var breake bool
var outPath string
var skipZz bool

func newGentestCmd() *cobra.Command {
	gentestCmd := &cobra.Command{
		Use:   "gentest",
		Short: "generate test sqls by zz and yy",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if yyPath == "" {
				return errors.New("yy are required")
			}

			return nil
		},
		Run: gentestAction,
	}
	gentestCmd.Flags().BoolVarP(&format, "format", "F", true,
		"generate sql that is convenient for reading(not implement yet)")
	gentestCmd.Flags().BoolVarP(&breake, "break", "B", false,
		"break zz yy result to two resource")
	gentestCmd.Flags().StringVarP(&outPath, "output", "O","output",
		"sql output file path")
	gentestCmd.Flags().BoolVar(&skipZz, "skip-zz", false,
		"skip gen data phase, only use yy to generate random sqls")

	return gentestCmd
}

// generate all sqls and write them into file
func gentestAction(cmd *cobra.Command, args []string) {
	ddls, randomSqls := getSqls()

	if breake {
		if !skipZz {
			err := ioutil.WriteFile(outPath+".data.sql",
				[]byte(strings.Join(ddls, ";\n") + ";"), os.ModePerm)
			if err != nil {
				log.Printf("write ddl in dist fail, %v\n", err)
			}
		}

		err := ioutil.WriteFile(outPath+".rand.sql",
			[]byte(strings.Join(randomSqls, ";\n") + ";"), os.ModePerm)
		if err != nil {
			log.Printf("write random sql in dist fail, %v\n", err)
		}
	} else {
		allSqls := make([]string, 0)
		allSqls = append(allSqls, ddls...)
		allSqls = append(allSqls, randomSqls...)

		err := ioutil.WriteFile(outPath + ".sql",
			[]byte(strings.Join(allSqls, ";\n") + ";"), os.ModePerm)
		if err != nil {
			log.Printf("sql output error, %v\n", err)
		}
	}
}