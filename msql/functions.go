package msql

import (
	"context"
	"errors"
	"fmt"
	"github.com/gomelon/melon/data"
	"github.com/gomelon/melon/data/engine"
	"github.com/gomelon/melon/data/query"
	"github.com/gomelon/melon/third_party/sqlx"
	"github.com/gomelon/meta"
	"github.com/gomelon/meta-templates/msql/parser"
	"github.com/huandu/xstrings"
	"go/types"
	"strings"
)

type functions struct {
	ruleParser     *data.RuleParser
	packageParser  *meta.PackageParser
	metaParser     *meta.MetaParser
	importTracker  meta.ImportTracker
	defaultDialect string
}

func NewFunctions(generator *meta.TemplateGenerator) *functions {
	useEngines()
	return &functions{
		ruleParser:     data.NewRuleParser(),
		packageParser:  generator.PackageParser(),
		metaParser:     generator.MetaParser(),
		importTracker:  generator.ImportTracker(),
		defaultDialect: "mysql",
	}
}

func useEngines() {
	engine.UseMySQL()
}

func (f *functions) FuncMap() map[string]any {
	return map[string]any{
		"rewriteSelectStmt": f.RewriteSelectStmt,
		"queryArgs":         f.QueryArgs,
		"directive":         f.Directive,
		"scanFields":        f.ScanFields,
		"selectMeta":        f.SelectMeta,
	}
}

func (f *functions) Directive(method types.Object) (directive string, err error) {
	directive, _, err = f.subjectMeta(method)
	if err != nil {
		return
	}

	if len(directive) > 0 {
		return
	}

	subject, err := f.ruleParser.ParseSubject(method.Name())
	if err != nil {
		return "", err
	}
	switch subject {
	case query.SubjectFind, query.SubjectCount, query.SubjectExists:
		directive = MetaSqlSelect
	case query.SubjectDelete:
		directive = MetaSqlDelete
	}
	return
}

func (f *functions) SelectMeta(tableMeta *Table, method types.Object) (selectMeta *Select, err error) {
	directive, selectMetaGroup, err := f.subjectMeta(method)
	if err != nil {
		return
	}

	if len(directive) > 0 && directive != MetaSqlSelect {
		return
	}

	if selectMetaGroup != nil {
		originSelectMeta := selectMetaGroup[0].(*Select)
		if len(originSelectMeta.Query) > 0 {
			selectMeta = originSelectMeta
			return
		}
		selectMeta = &Select{
			Query:     originSelectMeta.Query,
			Master:    originSelectMeta.Master,
			Omitempty: originSelectMeta.Omitempty,
		}
	}

	parsedQuery, err := f.ruleParser.Parse(method.Name())

	if parsedQuery == nil ||
		(parsedQuery.Subject() != query.SubjectFind &&
			parsedQuery.Subject() != query.SubjectCount &&
			parsedQuery.Subject() != query.SubjectExists) {
		if selectMeta != nil {
			err = fmt.Errorf("can not parse method to query,method=%s, possible reasons is %w",
				method.String(), err)
		} else {
			err = nil
		}
		return
	}

	parsedQuery = parsedQuery.With(query.WithTable(query.NewTable(tableMeta.Name)))
	if parsedQuery.FilterGroup() != nil {
		otherParams := f.packageParser.Params(method)[1:]
		namedArgs := make([]string, 0, len(otherParams))
		for _, param := range otherParams {
			namedArgs = append(namedArgs, param.Name())
		}
		err = parsedQuery.FilterGroup().FillNamedArgs(namedArgs)
		if err != nil {
			return
		}
	}

	sql, err := f.translateQuery(tableMeta, parsedQuery)
	if err != nil {
		return nil, err
	}

	if selectMeta == nil {
		selectMeta = &Select{
			Query:     sql,
			Master:    false,
			Omitempty: false,
		}
	} else {
		selectMeta.Query = sql
	}
	return
}

func (f *functions) RewriteSelectStmt(method types.Object, table *Table, sel *Select) (query string, err error) {
	dialect := f.dialect(table)

	query, _, err = f.compileNamedQuery(sel.Query, dialect)

	sqlParser, err := parser.New(dialect, query)
	if err != nil {
		err = fmt.Errorf("parse sql fail: %w", err)
		return
	}
	selectColumns, err := sqlParser.SelectColumns()
	if err != nil {
		err = fmt.Errorf("parse sql fail: %w", err)
		return
	}

	if len(selectColumns) == 1 && selectColumns[0].Alias == "*" {
		queryResultObject := f.packageParser.FirstResult(method)
		rowType := f.packageParser.UnderlyingType(queryResultObject.Type())
		rowStruct, ok := rowType.Underlying().(*types.Struct)
		if !ok {
			err = fmt.Errorf("parse sql fail: query result must a struct when select *, method=[%s],sql=%s",
				method.String(), sel.Query)
			return
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
	return
}

func (f *functions) ScanFields(method types.Object, table *Table, sql string, item string) (string, error) {
	dialect := f.dialect(table)

	var err error

	sqlParser, err := parser.New(dialect, sql)
	if err != nil {
		return "", fmt.Errorf("parse sql fail: %w, method=[%s],sql=%s", err, method.String(), sql)
	}
	columns, err := sqlParser.SelectColumns()
	if err != nil {
		return "", fmt.Errorf("parse sql fail: %w,method=[%s],sql=%s", err, method.String(), sql)
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
		return "", fmt.Errorf("parse sql fail:%w, method=[%s],sql=%s", err, method.String(), sql)
	}

	return result, nil
}

func (f *functions) QueryArgs(method types.Object, table *Table, sel *Select) (nameArgsStr string, err error) {
	dialect := f.dialect(table)
	_, queryNames, err := f.compileNamedQuery(sel.Query, dialect)
	if err != nil {
		return
	}

	methodParams := f.packageParser.Params(method)
	var toArgMethodParams []types.Object
	if len(methodParams) > 0 && f.packageParser.AssignableToCtx(methodParams[0].Type()) {
		toArgMethodParams = methodParams[1:]
	} else {
		toArgMethodParams = methodParams
	}

	if len(queryNames) == 0 {
		f.positionArgsStr(toArgMethodParams)
		return
	}

	nameArgsStr, err = f.nameArgsStr(queryNames, toArgMethodParams)
	if err != nil {
		err = fmt.Errorf("parse sql fail: %w, method=[%s],sql=%s", err, method.String(), sel.Query)
	}
	return
}

func (f *functions) positionArgsStr(toArgsMethodParams []types.Object) string {
	if len(toArgsMethodParams) == 0 {
		return ""
	}
	argsBuilder := strings.Builder{}
	argsBuilder.Grow(64)
	for _, queryName := range toArgsMethodParams {
		argsBuilder.WriteString(queryName.Name())
		argsBuilder.WriteRune(',')

	}
	return argsBuilder.String()
}

func (f *functions) nameArgsStr(queryNames []string, toArgsMethodParams []types.Object) (string, error) {
	if len(toArgsMethodParams) != len(queryNames) {
		err := fmt.Errorf("wrong number of args, want %d got %d", len(queryNames), len(toArgsMethodParams))
		return "", err
	}

	argsBuilder := strings.Builder{}
	argsBuilder.Grow(64)
	for _, queryName := range queryNames {
		argsBuilder.WriteString(queryName)
		argsBuilder.WriteRune(',')

	}
	return argsBuilder.String(), nil
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
			err = fmt.Errorf("msql: unsupported * mixed with specified fields query")
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

func (f *functions) connectTableQualifier(tableQualifier, column string) string {
	if len(tableQualifier) == 0 {
		return column
	}
	return tableQualifier + "." + column
}

func (f *functions) dialect(table *Table) string {
	dialect := table.Dialect
	if len(dialect) == 0 {
		dialect = f.defaultDialect
	}
	return strings.ToLower(dialect)
}

func (f *functions) translateQuery(tableMeta *Table, q *query.Query) (sql string, err error) {
	dialect := f.dialect(tableMeta)
	dialectEngine := engine.Engines[dialect]
	if dialectEngine == nil {
		err = fmt.Errorf("unsupported dialect,dialect=%s", dialect)
		return
	}
	translator := query.NewRDBTranslator(dialectEngine)
	return translator.Translate(context.Background(), q)
}

func (f *functions) compileNamedQuery(namedQuery, dialect string) (query string, names []string, err error) {
	bindType := sqlx.BindType(dialect)
	if bindType == 0 {
		err = fmt.Errorf("unsupported dialect,dialect=%s", dialect)
		return
	}
	query, names, err = sqlx.CompileNamedQuery([]byte(namedQuery), bindType)
	return
}

func (f *functions) subjectMeta(method types.Object) (directive string, group meta.Group, err error) {
	metaGroups := f.metaParser.ObjectMetaGroups(method, Directives...)
	if len(metaGroups) > 1 {
		err = fmt.Errorf("method can not use multiple %v,method=%s", Directives, method.String())
	} else if len(metaGroups) == 1 {
		for k, v := range metaGroups {
			directive = k
			group = v
			break
		}
	}
	return
}
