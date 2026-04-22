package sim

type Upgrader interface {
	Kind() string
	Apply(o *Orb)
}
