package sim

type ComponentKind string

const (
	KindInjector    ComponentKind = "injector"
	KindAccelerator ComponentKind = "accelerator"
	KindRotator     ComponentKind = "rotator"
)

type Component interface {
	Kind() ComponentKind
	Apply(s Subject) Subject
}
