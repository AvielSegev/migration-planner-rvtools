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
	}

	CreateTableStmt = `CREATE TABLE %s AS SELECT * FROM read_xlsx("%s",sheet="%s",all_varchar=true);`

	SelectOsStmt = `select "OS according to the VMware Tools" as name, count("OS according to the VMware Tools") as count from vinfo group by "OS according to the VMware Tools" order by "OS according to the VMware Tools";`

	SelectDatastoreStmt = `
WITH expanded AS (
       SELECT
           *,
           trim(unnest(string_split(hosts, ','))) AS ip
       FROM vdatastore
   )
   SELECT
       e."Free MiB"::double as freeCapacity,
       COALESCE(e."MHA", 'N/A') as mha,
       COALESCE(string_agg(h."Object ID", ', '), 'N/A') AS hosts,
       COALESCE(e."Type", 'N/A') as diskType,
       e."Capacity MiB"::double as totalCapacity,
       COALESCE(hba.Model, 'N/A') as hbaModel,
       COALESCE(hba.Type, 'N/A') as hbaType
   FROM expanded e
   LEFT JOIN vhost h ON h.Host = e.ip
   LEFT JOIN vhba hba ON hba.Host = e.ip
   GROUP BY ALL;
`

	SelectDatastoreSimpleStmt = `
   SELECT
       "Free MiB"::double as freeCapacity,
       COALESCE("MHA", 'N/A') as mha,
       'N/A' AS hosts,
       COALESCE("Type", 'N/A') as diskType,
       "Capacity MiB"::double as totalCapacity,
       'N/A' as hbaModel,
       'N/A' as hbaType
   FROM vdatastore;
`
)
