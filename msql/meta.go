package msql

import "github.com/gomelon/meta"

const (
	MetaSqlTable  = "sql:table"
	MetaSqlSelect = "sql:select"
	MetaSqlInsert = "sql:insert"
	MetaSqlUpdate = "sql:update"
	MetaSqlDelete = "sql:delete"
	MetaSqlNone   = "sql:none"
)

var (
	Metas      = []meta.Meta{&Table{}, &Select{}, &Insert{}, &Update{}, &Delete{}, &None{}}
	Directives = []string{MetaSqlSelect, MetaSqlInsert, MetaSqlUpdate, MetaSqlDelete, MetaSqlNone}
)

type Table struct {
	Name    string `json:"name"`
	Dialect string `json:"dialect"`
}

func (t *Table) PlaceAt() meta.Place {
	return meta.PlaceInterface
}

func (t *Table) Directive() string {
	return MetaSqlTable
}

func (t *Table) Repeatable() bool {
	return false
}

type Select struct {
	Query     string `json:"query,omitempty"`
	Master    bool   `json:"master,omitempty,string"`
	Omitempty bool   `json:"omitempty,omitempty,string"`
}

func (q *Select) PlaceAt() meta.Place {
	return meta.PlaceInterfaceMethod
}

func (q *Select) Directive() string {
	return MetaSqlSelect
}

func (q *Select) Repeatable() bool {
	return false
}

type Insert struct {
	Query     string `json:"query,omitempty"`
	Omitempty bool   `json:"omitempty,omitempty"`
}

func (q *Insert) PlaceAt() meta.Place {
	return meta.PlaceInterfaceMethod
}

func (q *Insert) Directive() string {
	return MetaSqlInsert
}

func (q *Insert) Repeatable() bool {
	return false
}

type Update struct {
	Query     string `json:"query,omitempty"`
	Omitempty bool   `json:"omitempty,omitempty"`
}

func (q *Update) PlaceAt() meta.Place {
	return meta.PlaceInterfaceMethod
}

func (q *Update) Directive() string {
	return MetaSqlUpdate
}

func (q *Update) Repeatable() bool {
	return false
}

type Delete struct {
	Query     string `json:"query,omitempty"`
	Omitempty bool   `json:"omitempty,omitempty"`
}

func (q *Delete) PlaceAt() meta.Place {
	return meta.PlaceInterfaceMethod
}

func (q *Delete) Directive() string {
	return MetaSqlDelete
}

func (q *Delete) Repeatable() bool {
	return false
}

type None struct {
}

func (q *None) PlaceAt() meta.Place {
	return meta.PlaceInterfaceMethod
}

func (q *None) Directive() string {
	return MetaSqlNone
}

func (q *None) Repeatable() bool {
	return false
}
