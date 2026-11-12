package main

import (
	"errors"
	"math"
	"strings"

	"github.com/pingcap/go-randgen/gendata"
	"github.com/spf13/cobra"
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
		RunE: gentestAction,
	}
	gentestCmd.Flags().BoolVarP(&format, "format", "F", true,
		"generate sql that is convenient for reading(not implement yet)")
	gentestCmd.Flags().BoolVarP(&breake, "break", "B", false,
		"break zz yy result to two resource")

	return gentestCmd
}

// generate all sqls and write them into file
func gentestAction(cmd *cobra.Command, args []string) error {
	var keyf gendata.Keyfun
	var ddls []string

	if !skipZz {
		ddls, keyf = getDdls()
	} else {
		keyf = gendata.NewKeyfun(nil, nil)
	}

	randomSqls := getRandSqls(keyf)
	if !breake {
		return writeSQLArtifacts(sqlArtifact{
			path: outPath + ".sql",
			sql:  strings.Join(append(ddls, randomSqls...), ";\n") + ";",
		})
	}

	random := randomSQLArtifact(randomSqls)
	if skipZz {
		return writeSQLArtifacts(random)
	}
	return writeSQLArtifacts(sqlArtifact{
		path: outPath + ".data.sql",
		sql:  strings.Join(ddls, ";\n") + ";",
	}, random)
}
