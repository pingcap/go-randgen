package generators

type Datetime struct {
	Date
	Time
}

func (d *Datetime) Gen() string {
	return d.Date.Gen() + " " + d.Time.Gen()
}
