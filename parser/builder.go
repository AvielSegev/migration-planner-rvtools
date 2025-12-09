package parser

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/ingest_rvtools.go.tmpl
var ingestRvtoolsTemplate string

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

// IngestRvtoolsQuery returns a query that creates all tables from an RVTools Excel file.
// The returned query contains %s placeholders for the Excel file path.
func (b *QueryBuilder) IngestRvtoolsQuery() string {
	return b.buildQuery("ingest_rvtools", ingestRvtoolsTemplate, nil)
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
		query, err := b.buildDatastoreQuery(ctx)
		if err != nil {
			return nil, fmt.Errorf("building datastore query: %w", err)
		}
		queries[Datastore] = query
	}

	if ctx.HasTable("vnetwork") {
		query, err := b.buildNetworkQuery(ctx)
		if err != nil {
			return nil, fmt.Errorf("building network query: %w", err)
		}
		queries[Network] = query
	}

	if ctx.HasTable("vhost") {
		queries[Host] = b.buildQuery("host_query", hostQueryTemplate, nil)
	}

	return queries, nil
}

type vmQueryParams struct {
	HasVInfo                bool
	HasVCpu                 bool
	HasVMemory              bool
	HasVDisk                bool
	HasVNetwork             bool
	FolderColumn            string
	NetworkColumns          string
	DiskPathColumn          string
	UUIDColumn              string
	HasSharingMode          bool
	HasDiskUUID             bool
	HasTotalDiskCapacityMiB bool
	HasProvisionedMiB       bool
	HasResourcePool         bool
}

func (b *QueryBuilder) buildVMQuery(ctx *SchemaContext) (string, error) {
	params := vmQueryParams{
		HasVInfo:                true,
		HasVCpu:                 ctx.HasTable("vcpu"),
		HasVMemory:              ctx.HasTable("vmemory"),
		HasVDisk:                ctx.HasTable("vdisk"),
		HasVNetwork:             ctx.HasTable("vnetwork"),
		FolderColumn:            "Folder ID",
		DiskPathColumn:          "Path",
		UUIDColumn:              "SMBIOS UUID",
		HasTotalDiskCapacityMiB: ctx.HasColumn("vinfo", "Total disk capacity MiB"),
		HasProvisionedMiB:       ctx.HasColumn("vinfo", "Provisioned MiB"),
		HasResourcePool:         ctx.HasColumn("vinfo", "Resource pool"),
	}

	if ctx.HasColumn("vinfo", "Folder") && !ctx.HasColumn("vinfo", "Folder ID") {
		params.FolderColumn = "Folder"
	}
	if ctx.HasColumn("vinfo", "VM UUID") && !ctx.HasColumn("vinfo", "SMBIOS UUID") {
		params.UUIDColumn = "VM UUID"
	}

	networkCols := ctx.GetColumnsLike("vinfo", "Network #")
	if len(networkCols) == 0 {
		params.NetworkColumns = "NULL"
	} else {
		quoted := make([]string, len(networkCols))
		for i, col := range networkCols {
			quoted[i] = fmt.Sprintf(`i."%s"`, col)
		}
		params.NetworkColumns = strings.Join(quoted, ", ")
	}

	if params.HasVDisk {
		if ctx.HasColumn("vdisk", "Disk Path") && !ctx.HasColumn("vdisk", "Path") {
			params.DiskPathColumn = "Disk Path"
		}
		params.HasSharingMode = ctx.HasColumn("vdisk", "Sharing mode")
		params.HasDiskUUID = ctx.HasColumn("vdisk", "Disk UUID")
	}

	return b.buildQuery("vm_query", vmQueryTemplate, params), nil
}

type datastoreQueryParams struct {
	HasVHBA bool
}

func (b *QueryBuilder) buildDatastoreQuery(ctx *SchemaContext) (string, error) {
	params := datastoreQueryParams{
		HasVHBA: ctx.HasTable("vhba"),
	}
	return b.buildQuery("datastore_query", datastoreQueryTemplate, params), nil
}

type networkQueryParams struct {
	HasDvPort bool
}

func (b *QueryBuilder) buildNetworkQuery(ctx *SchemaContext) (string, error) {
	params := networkQueryParams{
		HasDvPort: ctx.HasTable("dvport"),
	}
	return b.buildQuery("network_query", networkQueryTemplate, params), nil
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
