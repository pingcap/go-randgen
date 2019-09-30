package main

import (
	"errors"
	"github.com/dqinyuan/go-randgen/gendata"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
)

var format bool
var breake bool

func newGentestCmd() *cobra.Command {
	gentestCmd := &cobra.Command{
		Use:   "gentest",
		Short: "generate test sqls by zz and yy",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if yyPath == "" {
				return errors.New("yy are required")
			}

			if queries < 0 {
				return errors.New("queries num must be positive or zero")
			}

			if maxRecursive <= 0 {
				maxRecursive = math.MaxInt32
			}

			return nil
		},
		Run: gentestAction,
	}
	gentestCmd.Flags().BoolVarP(&format, "format", "F", true,
		"generate sql that is convenient for reading(not implement yet)")
	gentestCmd.Flags().BoolVarP(&breake, "break", "B", false,
		"break zz yy result to two resource")

	return gentestCmd
}

// generate all sqls and write them into file
func gentestAction(cmd *cobra.Command, args []string) {

	var keyf gendata.Keyfun
	var ddls []string

	if !skipZz {
		ddls, keyf = getDdls()
	} else {
		keyf = gendata.NewKeyfun(nil, nil)
	}

	randomSqls := getRandSqls(keyf)

	if breake {
		if !skipZz {
			err := ioutil.WriteFile(outPath+".data.sql",
				[]byte(strings.Join(ddls, ";\n") + ";"), os.ModePerm)
			if err != nil {
				log.Printf("write ddl in dist fail, %v\n", err)
			}
		}

		dumpRandSqls(randomSqls)
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