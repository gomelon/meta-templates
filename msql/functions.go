package msql

import (
	"errors"
	"fmt"
	"github.com/gomelon/meta"
	"github.com/gomelon/meta-templates/msql/parser"
	"github.com/huandu/xstrings"
	"go/types"
	"strings"
)

type functions struct {
	packageParser  *meta.PackageParser
	metaParser     *meta.MetaParser
	importTracker  meta.ImportTracker
	defaultDialect string
}

func NewFunctions(generator *meta.TemplateGenerator) *functions {
	return &functions{
		packageParser:  generator.PackageParser(),
		metaParser:     generator.MetaParser(),
		importTracker:  generator.ImportTracker(),
		defaultDialect: "mysql",
	}
}

func (f *functions) FuncMap() map[string]any {
	return map[string]any{
		"rewriteSelectStmt": f.RewriteSelectStmt,
		"nameArgs":          f.NameArgs,
		"scanFields":        f.ScanFields,
	}
}

func (f *functions) RewriteSelectStmt(method types.Object, table *Table, sel *Select) string {
	dialect := table.Dialect
	if len(dialect) == 0 {
		dialect = f.defaultDialect
	}

	query := sel.Query
	sqlParser, err := parser.New(dialect, query)
	if err != nil {
		panic(fmt.Errorf("parse sql fail: %w", err))
	}
	selectColumns, err := sqlParser.SelectColumns()
	if err != nil {
		panic(fmt.Errorf("parse sql fail: %w", err))
	}

	if len(selectColumns) == 1 && selectColumns[0].Alias == "*" {
		queryResultObject := f.packageParser.FirstResult(method)
		rowType := f.packageParser.UnderlyingType(queryResultObject.Type())
		rowStruct, ok := rowType.Underlying().(*types.Struct)
		if !ok {
			panic(fmt.Errorf("query result must a struct when select *, method=[%s],sql=%s",
				method.String(), sel.Query))
		}

		column := selectColumns[0]
		numFields := rowStruct.NumFields()
		columnNames := make([]string, 0, numFields)
		for i := 0; i < numFields; i++ {
			columnName := xstrings.ToSnakeCase(rowStruct.Field(i).Name())
			columnNames = append(columnNames, f.connectTableQualifier(column.TableQualifier, columnName))
		}

		qualifierStarStr := f.connectTableQualifier(column.TableQualifier, "*")
		selectColumnStr := strings.Join(columnNames, ", ")
		query = strings.Replace(query, qualifierStarStr, selectColumnStr, 1)
	}
	return query
}

func (f *functions) ScanFields(method types.Object, table *Table, sel *Select, item string) string {
	dialect := table.Dialect
	if len(dialect) == 0 {
		dialect = f.defaultDialect
	}

	var err error

	query := sel.Query
	sqlParser, err := parser.New(dialect, query)
	if err != nil {
		panic(fmt.Errorf("parse sql fail: %w, method=[%s],sql=%s",
			err, method.String(), sel.Query))
	}
	columns, err := sqlParser.SelectColumns()
	if err != nil {
		panic(fmt.Errorf("parse sql fail: %w", err))
	}

	queryResultObject := f.packageParser.FirstResult(method)
	rowType := f.packageParser.UnderlyingType(queryResultObject.Type())

	var result string
	switch rowType := rowType.(type) {
	case *types.Struct:
		result, err = f.scanFieldsForStruct(rowType, columns, item)
	case *types.Basic:
		result, err = f.scanFieldsForBasic(rowType, columns, item)
	}

	if err != nil {
		panic(fmt.Errorf("parse sql fail:%w, method=[%s],sql=%s",
			err, method.String(), sel.Query))
	}

	return result
}

func (f *functions) scanFieldsForBasic(rowType *types.Basic, columns []*parser.Column,
	item string) (string, error) {
	if len(columns) > 1 || columns[0].Alias == "*" {
		return "", errors.New("when the query result is a basic type, select must be a specified field")
	}
	return "&" + item, nil
}

func (f *functions) scanFieldsForStruct(rowType *types.Struct, columns []*parser.Column,
	item string) (string, error) {
	if len(columns) == 1 && columns[0].Alias == "*" {
		return f.scanFieldsForStar(rowType, item)
	} else {
		return f.scanFieldsForMultipleColumn(rowType, columns, item)
	}

}

func (f *functions) scanFieldsForStar(rowType *types.Struct, item string) (string, error) {
	numFields := rowType.NumFields()
	toScanFieldNames := make([]string, 0, numFields)
	for i := 0; i < numFields; i++ {
		fieldName := rowType.Field(i).Name()
		toScanFieldName := "&" + item + "." + fieldName
		toScanFieldNames = append(toScanFieldNames, toScanFieldName)
	}
	return strings.Join(toScanFieldNames, ", "), nil
}

func (f *functions) scanFieldsForMultipleColumn(rowType *types.Struct, columns []*parser.Column,
	item string) (result string, err error) {

	toScanFieldNames := make([]string, 0, len(columns))
	structFieldNames := make(map[string]bool, rowType.NumFields())
	for i := 0; i < rowType.NumFields(); i++ {
		structFieldNames[rowType.Field(i).Name()] = true
	}
	for _, column := range columns {
		if column.Alias == "*" {
			err = fmt.Errorf("msql: unsupported * mixed with specified fields")
			return
		}
		fieldName := xstrings.ToCamelCase(column.Alias)
		if !structFieldNames[fieldName] {
			err = fmt.Errorf("msql: can't find field name in struct, field=%s,rowType=%s",
				fieldName, rowType.String())
		}

		toScanFieldName := "&" + item + "." + fieldName
		toScanFieldNames = append(toScanFieldNames, toScanFieldName)
	}
	result = strings.Join(toScanFieldNames, ", ")
	return
}

func (f *functions) NameArgs(method types.Object) string {
	sqlPkg := f.importTracker.Import("database/sql")
	otherParams := f.packageParser.Params(method)[1:]
	nameArgs := strings.Builder{}

	for i, param := range otherParams {
		nameArgs.WriteString(sqlPkg)
		nameArgs.WriteString(".Named(\"")
		nameArgs.WriteString(param.Name())
		nameArgs.WriteString("\",")
		nameArgs.WriteString(param.Name())
		nameArgs.WriteString("), ")
		if (i+1)%3 == 0 {
			nameArgs.WriteRune('\n')
		}
	}
	return nameArgs.String()
}

func (f *functions) connectTableQualifier(tableQualifier, column string) string {
	if len(tableQualifier) == 0 {
		return column
	}
	return tableQualifier + "." + column
}
