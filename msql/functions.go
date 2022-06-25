package msql

import (
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
		column := selectColumns[0]
		rowStruct, ok := rowType.Underlying().(*types.Struct)
		if !ok {
			panic(fmt.Errorf("query result must a struct when select *"))
		}

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

func (f *functions) ScanFields(method types.Object, table *Table, sel *Select, po string) string {
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

	if len(selectColumns) == 1 {
		if selectColumns[0].Alias == "*" {
			queryResultObject := f.packageParser.FirstResult(method)
			rowType := f.packageParser.UnderlyingType(queryResultObject.Type())
			rowStruct, ok := rowType.Underlying().(*types.Struct)
			if !ok {
				panic(fmt.Errorf("query result must a struct when select *"))
			}

			numFields := rowStruct.NumFields()
			toScanFieldNames := make([]string, 0, numFields)
			for i := 0; i < numFields; i++ {
				fieldName := rowStruct.Field(i).Name()
				toScanFieldName := "&" + po + "." + fieldName
				toScanFieldNames = append(toScanFieldNames, toScanFieldName)
			}
			return strings.Join(toScanFieldNames, ", ")
		} else {
			return "&" + po
		}
	}
	//TODO 非select * 的场景
	return query
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
