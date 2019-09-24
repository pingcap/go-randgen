package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dqinyuan/go-randgen/compare"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var dsn1 string
var dsn2 string
var nonOrder bool
var dumpDir string

func newExecCmd() *cobra.Command {
	execCmd := &cobra.Command{
		Use:   "exec",
		Short: "exec sql in two dsn and compare their result",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if yyPath == "" {
				return errors.New("yy are required")
			}
			if dsn1 == "" || dsn2 == "" {
				return errors.New("dsn must have a pair")
			}

			return nil
		},
		Run: execAction,
	}

	execCmd.Flags().StringVar(&dsn1, "dsn1", "", "one of compare mysql dsn")
	execCmd.Flags().StringVar(&dsn2, "dsn2", "", "another compare mysql dsn")
	execCmd.Flags().BoolVar(&nonOrder, "non-order",
		true, "compare sql result without order")
	execCmd.Flags().StringVar(&dumpDir, "dump",
		"dump", "inconsistent sqls dump directory")

	return execCmd
}

type dumpInfo struct {
	num int   // 编号
	sql string
	dsn1 string
	dsn2 string
	dsn1Res *compare.DsnRes
	dsn2Res *compare.DsnRes
}

func (dump *dumpInfo) String() string {
	bs := &bytes.Buffer{}
	dsn1Tag := fmt.Sprintf("[[%s]]\n\n", dump.dsn1)
	dsn2Tag := fmt.Sprintf("[[%s]]\n\n", dump.dsn2)

	// [sql]
	bs.WriteString("[sql]\n\n")
	bs.WriteString(dump.sql + "\n\n")

	// [err]
	bs.WriteString("[err]\n\n")
	bs.WriteString(dsn1Tag)
	if dump.dsn1Res.Err != nil {
		bs.WriteString(dump.dsn1Res.Err.Error())
	}
	bs.WriteString(dsn2Tag)
	if dump.dsn2Res.Err != nil {
		bs.WriteString(dump.dsn2Res.Err.Error())
	}

	// [compare]
	bs.WriteString("[compare]\n\n")
	dsn1Colored, dsn2Colored := getColorDiff(dump.dsn1Res.Res.String(),
		dump.dsn2Res.Res.String())
	bs.WriteString(dsn1Tag)
	bs.WriteString(dsn1Colored + "\n\n")
	bs.WriteString(dsn2Tag)
	bs.WriteString(dsn2Colored)

	return bs.String()
}

// dump inconsistent sqls and diff info into dump dir
func dumpVisitor(dsn1, dsn2 string) compare.Visitor {
	count := 0
	return func(sql string, dsn1Res *compare.DsnRes, dsn2Res *compare.DsnRes) error {

		info := &dumpInfo{
			num:count,
			sql:sql,
			dsn1:dsn1,
			dsn2:dsn2,
			dsn1Res:dsn1Res,
			dsn2Res:dsn2Res,
		}

		err := ioutil.WriteFile(filepath.Join(dumpDir,
			fmt.Sprintf("%d.log", count)), []byte(info.String()), os.ModePerm)
		if err != nil {
			return err
		}
		count++
		return nil
	}
}

func execAction(cmd *cobra.Command, args []string) {
	// compare two dsn
	ddls, randSqls := getSqls()

	db1, err := compare.OpenDBWithRetry("mysql", dsn1)
	if err != nil {
		log.Fatalf("connect dsn1 %s error %v\n", dsn1, err)
	}

	db2, err := compare.OpenDBWithRetry("mysql", dsn2)
	if err != nil {
		log.Fatalf("connect dsn2 %s error %v\n", dsn2, err)
	}

	// ddls must exec without error
	errSql, err := compare.ExecSqlsInDbs(ddls, db1, db2)
	if err != nil {
		log.Printf("Fatal Error: data prepar ddl exec error %v\n", err)
		log.Fatalln(errSql)
	}

	if isDirExist(dumpDir) {
		log.Fatalln("Fatal Error: dump directory already exist")
	}

	err = os.MkdirAll(dumpDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Fatal Error: dump dir %s create fail %v\n", dumpDir, err)
	}

	err = compare.ByDb(randSqls, db1, db2, nonOrder,
		dumpVisitor(dsn1, dsn2))

	if err != nil {
		log.Fatalf("Fatal Error: dump fail  %v\n", err)
	}

	log.Println("dump ok")
}

func isDirExist(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// delete with red and insert with green
// res1 edit path to res2
func getColorDiff(res1, res2 string) (string, string) {
	greenColor := color.New(color.FgGreen)
	greenColor.EnableColor()
	green := greenColor.SprintFunc()
	redColor := color.New(color.FgRed)
	redColor.EnableColor()
	red := redColor.SprintFunc()
	patch := diffmatchpatch.New()
	diff := patch.DiffMain(res1, res2, false)
	var res1Buf, res2Buf bytes.Buffer
	for _, d := range diff {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			res1Buf.WriteString(d.Text)
			res2Buf.WriteString(d.Text)
		case diffmatchpatch.DiffDelete:
			res1Buf.WriteString(red(d.Text))
		case diffmatchpatch.DiffInsert:
			res2Buf.WriteString(green(d.Text))
		}
	}
	return res1Buf.String(), res2Buf.String()
}
