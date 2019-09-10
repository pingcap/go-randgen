package generators

import (
	"go-randgen/resource"
	"log"
	"math/rand"
	"runtime/debug"
	"strings"
)

type English struct {
	// english dict from resource/english.txt
	dict []string
}

func newEnglish() *English {
	enBytes, err := resource.Asset("resource/english.txt")
	if err != nil {
		log.Fatalf("english dict read error %v\n %s\n", err, debug.Stack())
	}

	englishDict := string(enBytes)
	englishs := strings.Split(englishDict, "\n")

	return &English{englishs}
}

func (e *English) Gen() string {
	return `"` + e.dict[rand.Intn(len(e.dict))] + `"`
}

