package definitions

var (
	Sheets = []string{
		"vInfo",
		"vCPU",
		"vMemory",
		"vDisk",
		"vPartition",
		"vHost",
		"vDatastore",
		"vNetwork",
		"vCluster",
		"dvSwitch",
		"vHBA",
		"dvPort",
	}

	CreateTableStmt = `CREATE TABLE %s AS SELECT * FROM read_xlsx("%s",sheet="%s",all_varchar=true);`

	SelectOsStmt = `SELECT
		"OS according to the VMware Tools" as "name",
		COUNT("OS according to the VMware Tools") as "count" from vinfo
		WHERE "OS according to the VMware Tools" IS NOT NULL
		GROUP BY "OS according to the VMware Tools"
		ORDER BY "OS according to the VMware Tools";
`

	SelectDatastoreStmt = `
WITH expanded AS (
       SELECT
           d.*,
           trim(unnest(string_split(d.Hosts, ','))) AS ip,
           regexp_extract(d."Address", 'vmhba[0-9]+') as hba_device
       FROM vdatastore d
       WHERE d."Hosts" IS NOT NULL
   ),
   with_host AS (
       SELECT DISTINCT
           vh."Cluster",
           e."Address",
           e."Name",
           e."Free MiB",
           e."MHA",
           e."Capacity MiB",
           e."Type",
           e.ip,
           vh."Object ID",
           e.hba_device
       FROM expanded e
       JOIN vhost vh ON vh.Host = e.ip
   ),
   with_hba AS (
       SELECT DISTINCT
           w."Cluster",
           w."Address",
           w."Name",
           w."Free MiB",
           w."MHA",
           w."Capacity MiB",
           w."Type",
           w.ip,
           w."Object ID",
           FIRST(hba."Type") OVER (PARTITION BY w."Name") as hba_type
       FROM with_host w
       LEFT JOIN vhba hba ON hba."Device" = w.hba_device
   )
   SELECT
       w."Cluster" as "cluster",
       COALESCE(w."Address", w."Name") as "diskId",
       (w."Free MiB"::double / 1024)::integer as "freeCapacityGB",
       (w."MHA" = 'True') as "hardwareAcceleratedMove",
       COALESCE(string_agg(DISTINCT w."Object ID", ', '), 'N/A') AS "hostId",
       'N/A' as "model",
       CASE
           WHEN w."Type" = 'NFS' THEN 'N/A'
           WHEN w."Address" LIKE 'naa.%' THEN 'iSCSI'
           WHEN w.hba_type IS NOT NULL THEN w.hba_type
           ELSE 'N/A'
       END as "protocolType",
       (w."Capacity MiB"::double / 1024)::integer as "totalCapacityGB",
       COALESCE(w."Type", 'N/A') as "type",
       'N/A' as "vendor"
   FROM with_hba w
   WHERE w."Cluster" IS NOT NULL
   GROUP BY w."Cluster", w."Address", w."Name", w."Free MiB", w."MHA", w."Capacity MiB", w."Type", w.hba_type;
`

	SelectDatastoreSimpleStmt = `
WITH expanded AS (
       SELECT
           d.*,
           trim(unnest(string_split(d.Hosts, ','))) AS ip
       FROM vdatastore d
       WHERE d."Hosts" IS NOT NULL
   )
   SELECT
       vh."Cluster" as "cluster",
       COALESCE(e."Address", e."Name") as "diskId",
       (e."Free MiB"::double / 1024)::integer as "freeCapacityGB",
       (e."MHA" = 'True') as "hardwareAcceleratedMove",
       COALESCE(string_agg(DISTINCT vh."Object ID", ', '), 'N/A') AS "hostId",
       'N/A' as "model",
       CASE
           WHEN e."Type" = 'NFS' THEN 'N/A'
           WHEN e."Address" LIKE 'naa.%' THEN 'iSCSI'
           ELSE 'N/A'
       END as "protocolType",
       (e."Capacity MiB"::double / 1024)::integer as "totalCapacityGB",
       COALESCE(e."Type", 'N/A') as "type",
       'N/A' as "vendor"
   FROM expanded e
   JOIN vhost vh ON vh.Host = e.ip
   WHERE vh."Cluster" IS NOT NULL
   GROUP BY vh."Cluster", e."Address", e."Name", e."Free MiB", e."MHA", e."Capacity MiB", e."Type";
`

	SelectHostStmt = `
   SELECT
       "Cluster" as "cluster",
       "# Cores"::integer as "cpuCores",
       "# CPU"::integer as "cpuSockets",
       "Object ID" as "id",
       "# Memory"::integer as "memoryMB",
       COALESCE("Model", 'N/A') as "model",
       COALESCE("Vendor", 'N/A') as "vendor"
   FROM vhost
   WHERE "Cluster" IS NOT NULL;
`

	SelectNetworkStmt = `
   SELECT
       n."Cluster" as "cluster",
       COALESCE(n."Switch", '') as "dvswitch",
       n."Network" as "name",
       'distributed' as "type",
       COALESCE(p."VLAN", '') as "vlanId",
       COUNT(*)::integer as "vmsCount"
   FROM vnetwork n
   LEFT JOIN dvport p ON n."Network" = p."Port"
   WHERE n."Cluster" IS NOT NULL
   GROUP BY n."Cluster", n."Switch", n."Network", p."VLAN";
`

	SelectNetworkSimpleStmt = `
   SELECT
       "Cluster" as "cluster",
       COALESCE("Switch", '') as "dvswitch",
       "Network" as "name",
       'distributed' as "type",
       '' as "vlanId",
       COUNT(*)::integer as "vmsCount"
   FROM vnetwork
   WHERE "Cluster" IS NOT NULL
   GROUP BY "Cluster", "Switch", "Network";
`
)
