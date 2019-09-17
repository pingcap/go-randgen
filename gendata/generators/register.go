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

func Traverse(h func(string, Generator)) {
	for k, v := range gmap {
		h(k, v)
	}
}

func init() {
	gmap["digit"] = &Digit{}
	gmap["letter"] = &Letter{}
	/*  temporal
    yyyy-MM-dd HH:mm:ss.SSS
    */
	gmap["date"] = newTemporal(yyyy, dd)
	gmap["year"] = newTemporal(yyyy, yyyy)
	gmap["month"] = newTemporal(MM, MM)
	gmap["day"] = newTemporal(dd, dd)
	gmap["hour"] = newTemporal(HH, HH)
	gmap["minute"] = newTemporal(mm, mm)
	gmap["second"] = newTemporal(ss, ss)
	gmap["microsecond"] = newTemporal(SSS, SSS)
	gmap["time"] = newTemporal(HH, ss)
	gmap["datetime"] = newTemporal(yyyy, ss)
	gmap["second_microsecond"] = newTemporal(ss, SSS)
	gmap["minute_microsecond"] = newTemporal(mm, SSS)
	gmap["minute_second"] = newTemporal(mm, ss)
	gmap["hour_microsecond"] = newTemporal(HH, SSS)
	gmap["hour_second"] = newTemporal(HH, ss)
	gmap["hour_minute"] = newTemporal(HH, mm)
	gmap["day_microsecond"] = newTemporal(dd, SSS)
	gmap["day_second"] = newTemporal(dd, ss)
	gmap["day_minute"] = newTemporal(dd, mm)
	gmap["day_hour"] = newTemporal(dd, HH)
	gmap["year_month"] = newTemporal(yyyy, MM)


	gmap["timestamp"] = &Timestamp{}
	gmap["english"] = newEnglish()
	gmap["char"] = NewChar(10)
	gmap["bool"] = newInt(0, 1, "")
	gmap["boolean"] = newInt(0, 1, "")
	gmap["tinyint"] = newInt(-128, 127, "")
	gmap["tinyint_unsigned"] = newInt(0, 255, "")
	gmap["smallint"] = newInt(-32768, 32767, "")
	gmap["smallint_unsigned"] = newInt(0, 65535, "")
	gmap["mediumint"] = newInt(-8388608, 8388607, "")
	gmap["mediumint_unsigned"] = newInt(0, 16777215, "")
	gmap["int"] = newInt(0, -1, "")
	gmap["int_usigned"] = &Uint{}
	gmap["integer"] = newInt(0, -1, "")
	gmap["decimal"] = &Decimal{}
}
