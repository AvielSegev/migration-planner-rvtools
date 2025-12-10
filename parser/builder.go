package parser

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/create_schema.go.tmpl
var createSchemaTemplate string

//go:embed templates/ingest_rvtools.go.tmpl
var ingestRvtoolsTemplate string

//go:embed templates/ingest_sqlite.go.tmpl
var ingestSqliteTemplate string

//go:embed templates/vm_query.go.tmpl
var vmQueryTemplate string

//go:embed templates/datastore_query.go.tmpl
var datastoreQueryTemplate string

//go:embed templates/network_query.go.tmpl
var networkQueryTemplate string

//go:embed templates/os_query.go.tmpl
var osQueryTemplate string

//go:embed templates/host_query.go.tmpl
var hostQueryTemplate string

//go:embed templates/vcenter_query.go.tmpl
var vcenterQueryTemplate string

// Type represents the type of query to build
type Type int

const (
	VM Type = iota
	Datastore
	Network
	Host
	Os
	VCenter
)

func (q Type) String() string {
	switch q {
	case VM:
		return "vm"
	case Datastore:
		return "datastore"
	case Network:
		return "network"
	case Host:
		return "host"
	case Os:
		return "os"
	case VCenter:
		return "vcenter"
	default:
		return "unknown"
	}
}

// SchemaContext holds information about available tables and columns in the database.
type SchemaContext struct {
	Tables  map[string]bool            // table name -> exists
	Columns map[string]map[string]bool // table name -> column name -> exists
}

func (s *SchemaContext) HasTable(table string) bool {
	return s.Tables[table]
}

func (s *SchemaContext) HasColumn(table, column string) bool {
	if cols, ok := s.Columns[table]; ok {
		return cols[column]
	}
	return false
}

func (s *SchemaContext) GetColumnsLike(table, prefix string) []string {
	var result []string
	if cols, ok := s.Columns[table]; ok {
		for col := range cols {
			if strings.HasPrefix(col, prefix) {
				result = append(result, col)
			}
		}
	}
	sort.Strings(result)
	return result
}

// QueryBuilder builds SQL queries from templates.
type QueryBuilder struct{}

// NewBuilder creates a new Builder.
func NewBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

type ingestParams struct {
	FilePath string
}

// CreateSchemaQuery returns queries to create all RVTools tables with proper schema.
func (b *QueryBuilder) CreateSchemaQuery() string {
	return b.buildQuery("create_schema", createSchemaTemplate, nil)
}

// IngestRvtoolsQuery returns a query that inserts data from an RVTools Excel file into schema tables.
func (b *QueryBuilder) IngestRvtoolsQuery(filePath string) string {
	return b.buildQuery("ingest_rvtools", ingestRvtoolsTemplate, ingestParams{FilePath: filePath})
}

// IngestSqliteQuery returns a query that creates RVTools-shaped tables from a forklift SQLite database.
func (b *QueryBuilder) IngestSqliteQuery(filePath string) string {
	return b.buildQuery("ingest_sqlite", ingestSqliteTemplate, ingestParams{FilePath: filePath})
}

// Build generates all SQL queries based on the schema context.
func (b *QueryBuilder) Build(ctx *SchemaContext) (map[Type]string, error) {
	queries := make(map[Type]string)

	if ctx.HasTable("vinfo") {
		query, err := b.buildVMQuery(ctx)
		if err != nil {
			return nil, fmt.Errorf("building VM query: %w", err)
		}
		queries[VM] = query
		queries[Os] = b.buildQuery("os_query", osQueryTemplate, nil)
		queries[VCenter] = b.buildQuery("vcenter_query", vcenterQueryTemplate, nil)
	}

	if ctx.HasTable("vdatastore") && ctx.HasTable("vhost") {
		queries[Datastore] = b.buildDatastoreQuery()
	}

	if ctx.HasTable("vnetwork") {
		queries[Network] = b.buildNetworkQuery()
	}

	if ctx.HasTable("vhost") {
		queries[Host] = b.buildQuery("host_query", hostQueryTemplate, nil)
	}

	return queries, nil
}

type vmQueryParams struct {
	NetworkColumns string
}

func (b *QueryBuilder) buildVMQuery(ctx *SchemaContext) (string, error) {
	networkCols := ctx.GetColumnsLike("vinfo", "Network #")
	var networkColumns string
	if len(networkCols) == 0 {
		networkColumns = "NULL"
	} else {
		quoted := make([]string, len(networkCols))
		for i, col := range networkCols {
			quoted[i] = fmt.Sprintf(`i."%s"`, col)
		}
		networkColumns = strings.Join(quoted, ", ")
	}

	return b.buildQuery("vm_query", vmQueryTemplate, vmQueryParams{NetworkColumns: networkColumns}), nil
}

func (b *QueryBuilder) buildDatastoreQuery() string {
	return b.buildQuery("datastore_query", datastoreQueryTemplate, nil)
}

func (b *QueryBuilder) buildNetworkQuery() string {
	return b.buildQuery("network_query", networkQueryTemplate, nil)
}

func (b *QueryBuilder) buildQuery(name, tmplContent string, params any) string {
	tmpl, err := template.New(name).Parse(tmplContent)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}
