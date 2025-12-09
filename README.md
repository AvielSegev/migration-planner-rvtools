# rvtools

Parses VMware inventory data from RVTools Excel exports or forklift SQLite databases and outputs JSON.

## Usage

```
./rvtools -excel-file <path>
./rvtools -sqlite-file <path>
./rvtools -excel-file <path> -db-path <path>
```

- `-excel-file`: Path to RVTools Excel export (.xlsx)
- `-sqlite-file`: Path to forklift SQLite database
- `-db-path`: Optional path to persist DuckDB database to disk

One of `-excel-file` or `-sqlite-file` is required.

## Build

```
make build
make run EXCEL_FILE=path/to/file.xlsx
make clean
```

## Architecture

Uses DuckDB as the query engine. Both data sources are ingested into RVTools-shaped tables (vinfo, vcpu, vmemory, vdisk, vnetwork, vhost, vdatastore, dvport, vhba), then the same SQL query templates extract the inventory data.

### Excel ingestion
Direct table creation from Excel sheets using DuckDB's `read_xlsx()`.

### SQLite ingestion
Transforms the normalized forklift/vsphere model into flat RVTools tables:
- VM.NICs and VM.Disks JSON arrays are unnested into separate rows
- Cluster is derived via VM.Host -> Host.Cluster
- Datacenter is derived via Cluster.Parent -> Folder.Datacenter
- Datastore cluster mapping derived via Host.Datastores JSON array
