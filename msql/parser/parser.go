package parser

import (
	"fmt"
	"strings"
)

type Type int

const (
	TypeSelect Type = iota
	TypeInsert
	TypeUpdate
	TypeDelete
)

type Parser interface {
	Type() (Type, error)
	SelectColumns() ([]*Column, error)
}

func New(dialect string, sql string) (p Parser, err error) {
	dialectUpper := strings.ToUpper(dialect)
	switch dialectUpper {
	case "MYSQL":
		p, err = NewMySQL(sql)
	default:
		err = fmt.Errorf("sql parser: unsupported dialect %s", dialect)
	}
	return
}

type Column struct {
	Alias          string
	TableQualifier string
}
