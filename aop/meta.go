package aop

import "github.com/gomelon/meta"

const (
	MetaAopInterface = "aop:interface"
)

func AllMetas() []meta.Meta {
	return []meta.Meta{&Interface{}}
}

type Interface struct {
}

func (i *Interface) PlaceAt() meta.Place {
	return meta.PlaceStruct
}

func (i *Interface) Directive() string {
	return MetaAopInterface
}

func (i *Interface) Repeatable() bool {
	return false
}
