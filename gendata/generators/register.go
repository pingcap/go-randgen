package generators

// the implementation of it should not have status
type Generator interface {
	Gen() string
}

var gmap = make(map[string]Generator)

func Get(name string) Generator {
	g, ok := gmap[name]
	if !ok {
		return nil
	}
	return g
}

func init() {
	gmap["digit"] = &Digit{}
	gmap["letter"] = &Letter{}
	gmap["date"] = &Date{}
	gmap["year"] = &Year{}
	gmap["time"] = &Time{}
	gmap["datetime"] = &Datetime{}
	gmap["timestamp"] = &Timestamp{}
	gmap["english"] = newEnglish()
	gmap["bool"] = newInt(0, 1)
	gmap["boolean"] = newInt(0, 1)
	gmap["tinyint"] = newInt(-128, 127)
	gmap["smallint"] = newInt(-32768, 32767)
	gmap["mediumint"] = newInt(-8388608, 8388607)
	gmap["int"] = newInt(0, -1)
	gmap["integer"] = newInt(0, -1)
	gmap["decimal"] = &Decimal{}
}
