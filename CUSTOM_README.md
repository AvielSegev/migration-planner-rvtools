# Migration Planner RVTools Parser - Deep Dive Guide

## 📋 What Is This?

This is a **Go CLI tool** that extracts VMware vSphere inventory data and outputs it as **JSON**. It's designed to help with migration planning by providing a unified view of your VMware environment.

### The Problem It Solves

When planning migrations from VMware, you need detailed inventory data about:
- Virtual Machines (VMs)
- ESXi Hosts
- Datastores (storage)
- Networks
- Clusters

This data typically comes in two formats:
1. **RVTools Excel exports** - A popular VMware inventory tool that exports `.xlsx` files
2. **Forklift SQLite databases** - From the [Konveyor Forklift](https://github.com/kubev2v/forklift) migration toolkit

This tool **normalizes both formats** into a single JSON output structure.

---

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         INPUT SOURCES                           │
├─────────────────────────────┬───────────────────────────────────┤
│   RVTools Excel (.xlsx)     │   Forklift SQLite (.db)           │
│   └── Multiple sheets:      │   └── Normalized tables:          │
│       vInfo, vCPU, vMemory, │       VM, Host, Cluster,          │
│       vDisk, vNetwork...    │       Network, Datastore...       │
└─────────────────────────────┴───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DUCKDB (In-Memory)                           │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              RVTools-Shaped Schema Tables                 │  │
│  │  vinfo, vcpu, vmemory, vdisk, vnetwork, vhost,           │  │
│  │  vdatastore, dvport, vhba                                 │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│                              ▼                                   │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │              SQL Query Templates                           │  │
│  │  vm_query, host_query, datastore_query,                   │  │
│  │  network_query, os_query, vcenter_query                   │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       JSON OUTPUT                               │
│  {                                                              │
│    "vcenterId": "...",                                          │
│    "clusters": {                                                │
│      "ClusterName": {                                           │
│        "infra": { hosts, datastores, networks },                │
│        "vms": [ ... ]                                           │
│      }                                                          │
│    },                                                           │
│    "osSummary": [ { "name": "RHEL 8", "count": 42 }, ... ]      │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
```

### Key Insight: The "Two-Phase" Design

**Phase 1 - Ingestion:** Both input formats are transformed into a common set of "RVTools-shaped" tables in DuckDB. This is the normalization step.

**Phase 2 - Querying:** The same SQL query templates work against this common schema, regardless of the original data source.

---

## 📁 Directory Structure

```
migration-planner-rvtools/
├── main.go                    # CLI entry point
├── go.mod / go.sum            # Go dependencies
├── Makefile                   # Build commands
├── models/
│   └── models.go              # Go structs for inventory data
├── parser/
│   ├── parser.go              # Main parsing logic
│   ├── builder.go             # SQL query builder from templates
│   ├── parser_test.go         # Tests (Ginkgo/Gomega)
│   ├── parser_suite_test.go   # Test suite setup
│   └── templates/             # SQL templates (Go text/template)
│       ├── create_schema.go.tmpl     # Creates empty RVTools tables
│       ├── ingest_rvtools.go.tmpl    # Excel → RVTools tables
│       ├── ingest_sqlite.go.tmpl     # Forklift → RVTools tables
│       ├── vm_query.go.tmpl          # Query VMs
│       ├── host_query.go.tmpl        # Query hosts
│       ├── datastore_query.go.tmpl   # Query datastores
│       ├── network_query.go.tmpl     # Query networks
│       ├── os_query.go.tmpl          # Query OS distribution
│       └── vcenter_query.go.tmpl     # Query vCenter ID
│   └── testdata/              # Test fixtures
│       ├── fixtures.sql               # Full RVTools test data
│       ├── fixtures_incomplete.sql    # Partial data (edge cases)
│       └── create_forklift_sqlite.sql # Forklift schema + test data
└── rvtools                    # Compiled binary (after build)
```

---

## 🔧 How It Works

### 1. Entry Point (`main.go`)

```go
// CLI flags
-excel-file <path>     // RVTools Excel export
-sqlite-file <path>    // Forklift SQLite database
-db-path <path>        // Optional: persist DuckDB to disk
-enable-timing         // Print parsing duration
-debug                 // Enable debug logging
```

The main function:
1. Initializes DuckDB (in-memory or file-based)
2. Loads the `excel` extension for reading `.xlsx` files
3. Creates a `Parser` based on input type (RVTools or SQLite)
4. Calls `Parse()` and outputs JSON

### 2. Parser (`parser/parser.go`)

The `Parser` struct uses a **preprocessor pattern**:

```go
type Parser struct {
    db           *sql.DB       // DuckDB connection
    builder      *QueryBuilder // Generates SQL from templates
    preprocessor Preprocessor  // Either rvToolsPreprocessor or sqlitePreprocessor
}
```

**`Parse()` flow:**
1. `createSchema()` - Creates empty RVTools-shaped tables
2. `preprocessor.Process()` - Ingests data from the source
3. `builder.Build()` - Generates query SQL from templates
4. Reads all data: VMs, hosts, datastores, networks, OS summary, vCenter ID
5. `buildInventory()` - Groups everything by cluster name

### 3. SQL Templates (`parser/templates/`)

Templates use Go's `text/template` syntax. Key templates:

| Template | Purpose |
|----------|---------|
| `create_schema.go.tmpl` | Creates 9 empty tables with RVTools column names |
| `ingest_rvtools.go.tmpl` | Uses DuckDB's `read_xlsx()` to import Excel sheets |
| `ingest_sqlite.go.tmpl` | ATTACHes SQLite DB and transforms normalized data |
| `vm_query.go.tmpl` | JOINs vinfo + vcpu + vmemory + vdisk + vnetwork |

**Example - Excel ingestion:**
```sql
-- Uses DuckDB's Excel extension
INSERT INTO vinfo (...)
SELECT ... FROM read_xlsx('{{.FilePath}}', sheet='vInfo', all_varchar=true);
```

**Example - SQLite ingestion with JSON unnesting:**
```sql
-- Unnests JSON arrays (e.g., VM.Disks) into rows
INSERT INTO vdisk (...)
SELECT v.ID, disk->>'key', ...
FROM src.VM v,
LATERAL unnest(from_json(v.Disks, '[...]')) AS t(disk);
```

### 4. Data Models (`models/models.go`)

Key structs:

```go
// Top-level output
type Inventory struct {
    VcenterId string                   `json:"vcenterId"`
    Clusters  map[string]InventoryData `json:"clusters"`
    OsSummary []Os                     `json:"osSummary"`
}

// Per-cluster data
type InventoryData struct {
    Infra Infra `json:"infra"`  // hosts, datastores, networks
    VMs   []VM  `json:"vms"`
}

// VM with nested arrays
type VM struct {
    ID       string   `db:"VM ID"`
    Name     string   `db:"VM"`
    NICs     NICs     // Custom scanner for DuckDB LIST type
    Disks    Disks    // Custom scanner for DuckDB LIST type
    Networks Networks // List of network names
    // ... 30+ other fields
}
```

**Custom SQL Scanners:** `NICs`, `Disks`, and `Networks` implement `sql.Scanner` to handle DuckDB's native `LIST` type, converting it to Go slices.

---

## 🚀 Usage

### Build
```bash
make build
# Creates ./rvtools binary
```

### Run with RVTools Excel
```bash
./rvtools -excel-file /path/to/export.xlsx
```

### Run with Forklift SQLite
```bash
./rvtools -sqlite-file /path/to/forklift.db
```

### Persist DuckDB (for debugging)
```bash
./rvtools -excel-file export.xlsx -db-path ./debug.duckdb
# You can then open debug.duckdb with DuckDB CLI
```

### Run Tests
```bash
go test ./parser/... -v
# Uses Ginkgo/Gomega test framework
```

---

## 📊 Output JSON Structure

```json
{
  "vcenterId": "vcenter-uuid-001",
  "clusters": {
    "Production-Cluster": {
      "infra": {
        "hosts": [
          {
            "id": "host-001",
            "cpuCores": 16,
            "cpuSockets": 2,
            "memoryMB": 131072,
            "model": "ProLiant DL380 Gen10",
            "vendor": "HPE"
          }
        ],
        "datastores": [
          {
            "diskId": "datastore1",
            "freeCapacityGB": 512.0,
            "totalCapacityGB": 1024.0,
            "type": "VMFS"
          }
        ],
        "networks": [
          {
            "name": "VM Network",
            "vlanId": "341",
            "type": "distributed",
            "vmsCount": 15
          }
        ],
        "totalHosts": 3
      },
      "vms": [
        {
          "id": "vm-001",
          "name": "web-server-01",
          "host": "host-001",
          "powerState": "poweredOn",
          "cpuCount": 4,
          "memoryMB": 8192,
          "guestName": "Red Hat Enterprise Linux 8 (64-bit)",
          "disks": [
            {
              "key": "2000",
              "file": "[datastore1] web-server-01/disk1.vmdk",
              "capacity": 53687091200
            }
          ],
          "nics": [
            {
              "network": "VM Network",
              "mac": "00:50:56:aa:bb:01",
              "connected": true
            }
          ]
        }
      ]
    }
  },
  "osSummary": [
    { "name": "Red Hat Enterprise Linux 8 (64-bit)", "count": 42 },
    { "name": "Microsoft Windows Server 2019 (64-bit)", "count": 28 }
  ]
}
```

---

## 🔍 Key Technical Details

### DuckDB Choice
- **In-memory analytics database** optimized for OLAP queries
- Built-in Excel reader (`read_xlsx()`)
- Can attach SQLite databases directly
- Supports advanced SQL: `LATERAL`, `UNNEST`, JSON operators, `LIST` aggregates

### RVTools Excel Format
- Multiple sheets: vInfo, vCPU, vMemory, vDisk, vNetwork, vHost, vDatastore, dvPort, vHBA
- Uses "VM" as placeholder text in empty cells (handled with `TRY_CAST` and filters)
- Network columns: `Network #1` through `Network #25`

### Forklift SQLite Format
- Normalized relational model (VM → Host → Cluster → Folder → Datacenter)
- JSON columns for arrays: `VM.NICs`, `VM.Disks`, `Host.Datastores`
- Requires JOINs and JSON unnesting to flatten

### Data Flow for SQLite Ingestion
```
VM.Disks JSON array → LATERAL UNNEST → vdisk rows
VM.NICs JSON array  → LATERAL UNNEST → vnetwork rows
Host.Cluster ID     → JOIN Cluster   → Cluster name
Cluster.Parent      → JOIN Folder    → Folder.Datacenter → Datacenter name
```

---

## 🧪 Testing

Tests use **Ginkgo** (BDD-style) and **Gomega** (matchers):

```bash
# Run all tests
go test ./parser/... -v

# Run specific tests
go test ./parser/... -v --ginkgo.focus="SQLite ingestion"
```

**Test data:**
- `fixtures.sql` - Complete RVTools-shaped test data in DuckDB
- `fixtures_incomplete.sql` - Missing tables/columns (edge cases)
- `create_forklift_sqlite.sql` - Creates a SQLite DB with Forklift schema

---

## 🔗 Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/duckdb/duckdb-go/v2` | DuckDB Go driver |
| `github.com/georgysavva/scany/v2` | Struct scanning from SQL rows |
| `github.com/glebarez/go-sqlite` | Pure-Go SQLite (for tests) |
| `github.com/onsi/ginkgo/v2` | BDD test framework |
| `github.com/onsi/gomega` | Test matchers |
| `go.uber.org/zap` | Structured logging |

---

## 💡 Common Tasks

### Add a new field to VM output
1. Add field to `models.VM` struct
2. Update `vm_query.go.tmpl` SELECT clause
3. Update `parser.readVMs()` Scan arguments
4. If from a new table, update `ingest_*.go.tmpl` as needed

### Support a new RVTools sheet
1. Add table to `create_schema.go.tmpl`
2. Add INSERT statement to `ingest_rvtools.go.tmpl`
3. Create new query template if needed
4. Add to `parser.go` read and build functions

### Debug SQL queries
```bash
./rvtools -excel-file data.xlsx -db-path debug.duckdb -debug
# Then:
duckdb debug.duckdb
> SELECT * FROM vinfo LIMIT 5;
```

---

## ⚠️ Limitations & Notes

- **Network columns:** Supports up to 25 networks per VM (hardcoded limit)
- **Excel errors:** Missing sheets are silently ignored (graceful degradation)
- **Memory:** Large Excel files are loaded entirely into memory via DuckDB
- **SQLite only:** Forklift format is the only normalized format supported

---

## 🎯 Use Case: Migration Planning

This tool is part of a migration planning workflow:

1. **Collect inventory** from VMware using RVTools or Forklift
2. **Parse with this tool** to get normalized JSON
3. **Analyze** VMs for migration readiness (concerns, compatibility)
4. **Plan migration** waves based on clusters, workload dependencies

The JSON output can be consumed by:
- Migration planning UIs
- Automated analysis tools
- Capacity planning systems
- Documentation generators

