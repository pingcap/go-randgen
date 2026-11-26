package view

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pingcap/go-randgen/grammar"
	"github.com/pingcap/go-randgen/grammar/yacc_parser"
)

type ProductionView struct {
	Number int        `json:"number"`
	Head   string     `json:"head"`
	Alter  []*SeqView `json:"alter"`
}

type SeqView struct {
	Content string `json:"content"`
	Fanout  []int  `json:"fanout"`
}

type arraySet struct {
	arr []int
	set map[int]bool
}

func newArraySet() *arraySet {
	return &arraySet{arr: make([]int, 0), set: make(map[int]bool)}
}

func (a *arraySet) add(num int) {
	if !a.set[num] {
		a.arr = append(a.arr, num)
		a.set[num] = true
	}
}

func productionToJson(productions []*yacc_parser.Production,
	pMap map[string]*yacc_parser.Production) ([]byte, error) {
	pViews := make([]*ProductionView, 0)

	for _, production := range productions {
		seqs := production.Alter
		seqViews := make([]*SeqView, 0, len(seqs))
		seqSet := make(map[string]bool)
		for _, seq := range seqs {
			content := seq.String()
			if _, ok := seqSet[content]; ok {
				// compress the seq set
				continue
			}
			seqSet[content] = true

			fanout := newArraySet()
			for _, item := range seq.Items {
				if yacc_parser.NonTerminalInMap(pMap, item) {
					fanout.add(pMap[item.OriginString()].Number)
				}
			}

			seqViews = append(seqViews, &SeqView{Content: content, Fanout: fanout.arr})
		}

		pViews = append(pViews, &ProductionView{
			Number: production.Number,
			Head:   production.Head.OriginString(),
			Alter:  seqViews,
		})
	}

	jsonBytes, err := json.Marshal(pViews)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

func Graph(yy string, token ...string) (http.HandlerFunc, error) {
	if len(token) != 1 || token[0] == "" {
		return nil, fmt.Errorf("a graph access token is required")
	}
	authorization := sha256.Sum256([]byte("Bearer " + token[0]))
	_, productions, pMap, err := grammar.Parse(yy)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := productionToJson(productions, pMap)
	if err != nil {
		return nil, err
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		provided := sha256.Sum256([]byte(request.Header.Get("Authorization")))
		if subtle.ConstantTimeCompare(provided[:], authorization[:]) != 1 {
			writer.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		if request.Method != http.MethodGet {
			writer.Header().Set("Allow", http.MethodGet)
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writer.Header().Set("content-type", "application/json")
		writer.Write(jsonBytes)
	}, nil
}
