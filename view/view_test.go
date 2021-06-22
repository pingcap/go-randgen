package view

import (
	"testing"

	"github.com/pingcap/go-randgen/grammar"
	"github.com/stretchr/testify/assert"
)

func TestProductionToJson(t *testing.T) {

	_, productions, pMap, err := grammar.Parse(`
	aa:
	 bb mm CC dd mm | nn mm  dd g
	
	mm :
	 aa ll | nn
	
	g: HELLO | HELLO
	`)

	expected := `[{"number":0,"head":"aa","alter":[{"content":"bb mm CC dd mm","fanout":[1]},{"content":"nn mm dd g","fanout":[1,2]}]},{"number":1,"head":"mm","alter":[{"content":"aa ll","fanout":[0]},{"content":"nn","fanout":[]}]},{"number":2,"head":"g","alter":[{"content":"HELLO","fanout":[]}]}]`

	assert.Equal(t, nil, err)
	jsonBytes, err := productionToJson(productions, pMap)
	assert.Equal(t, nil, err)

	assert.Equal(t, expected, string(jsonBytes))
	//fmt.Println(string(jsonBytes))
}
