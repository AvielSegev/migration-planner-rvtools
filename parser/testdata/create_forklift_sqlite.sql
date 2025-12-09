-- Creates a forklift-shaped SQLite database for testing SQLite ingestion
-- Schema matches github.com/kubev2v/forklift/pkg/controller/provider/model/vsphere

CREATE TABLE About (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    APIVersion TEXT,
    Product TEXT,
    InstanceUuid TEXT
);

CREATE TABLE Datacenter (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Clusters TEXT,
    Networks TEXT,
    Datastores TEXT,
    Vms TEXT
);

CREATE TABLE Folder (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Datacenter TEXT,
    Folder TEXT,
    Children TEXT
);

CREATE TABLE Cluster (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Folder TEXT,
    Hosts TEXT,
    Networks TEXT,
    Datastores TEXT,
    DasEnabled INTEGER,
    DasVms TEXT,
    DrsEnabled INTEGER,
    DrsBehavior TEXT,
    DrsVms TEXT
);

CREATE TABLE Host (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Cluster TEXT,
    Status TEXT,
    InMaintenanceMode INTEGER,
    ManagementServerIp TEXT,
    Thumbprint TEXT,
    Timezone TEXT,
    CpuSockets INTEGER,
    CpuCores INTEGER,
    MemoryBytes INTEGER,
    ProductName TEXT,
    ProductVersion TEXT,
    Model TEXT,
    Vendor TEXT,
    Network TEXT,
    Networks TEXT,
    Datastores TEXT,
    HostScsiDisks TEXT,
    AdvancedOptions TEXT,
    HbaDiskInfo TEXT,
    HostScsiTopology TEXT
);

CREATE TABLE Network (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Tag TEXT,
    DVSwitch TEXT,
    Key TEXT,
    Host TEXT,
    VlanId TEXT
);

CREATE TABLE Datastore (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Type TEXT,
    Capacity INTEGER,
    Free INTEGER,
    MaintenanceMode TEXT,
    BackingDevicesNames TEXT
);

CREATE TABLE VM (
    ID TEXT PRIMARY KEY,
    Variant TEXT,
    Name TEXT,
    Parent TEXT,
    Revision INTEGER,
    Folder TEXT,
    Host TEXT,
    RevisionValidated INTEGER,
    PolicyVersion INTEGER,
    UUID TEXT,
    Firmware TEXT,
    PowerState TEXT,
    ConnectionState TEXT,
    CpuAffinity TEXT,
    CpuHotAddEnabled INTEGER,
    CpuHotRemoveEnabled INTEGER,
    MemoryHotAddEnabled INTEGER,
    FaultToleranceEnabled INTEGER,
    CpuCount INTEGER,
    CoresPerSocket INTEGER,
    MemoryMB INTEGER,
    GuestName TEXT,
    GuestNameFromVmwareTools TEXT,
    HostName TEXT,
    GuestID TEXT,
    BalloonedMemory INTEGER,
    IpAddress TEXT,
    NumaNodeAffinity TEXT,
    StorageUsed INTEGER,
    Snapshot TEXT,
    IsTemplate INTEGER,
    ChangeTrackingEnabled INTEGER,
    TpmEnabled INTEGER,
    Devices TEXT,
    NICs TEXT,
    Disks TEXT,
    Controllers TEXT,
    Networks TEXT,
    Concerns TEXT,
    GuestNetworks TEXT,
    GuestDisks TEXT,
    GuestIpStacks TEXT,
    SecureBoot INTEGER,
    ToolsStatus TEXT,
    ToolsRunningStatus TEXT,
    ToolsVersionStatus TEXT,
    DiskEnableUuid INTEGER,
    NestedHVEnabled INTEGER
);

-- Insert test data

INSERT INTO About VALUES (
    'about-1', '', 'vCenter', '{}', 1,
    '7.0', 'VMware vCenter Server', 'vcenter-uuid-001'
);

INSERT INTO Datacenter VALUES (
    'datacenter-1', '', 'TestDC', '{}', 1,
    '{}', '{}', '{}', '{}'
);

INSERT INTO Folder VALUES (
    'folder-1', '', 'host', '{"kind":"Datacenter","id":"datacenter-1"}', 1,
    'datacenter-1', '', '[]'
);

INSERT INTO Cluster VALUES (
    'cluster-1', '', 'TestCluster', '{"kind":"Folder","id":"folder-1"}', 1,
    'folder-1', '[]', '[]', '[]', 0, '[]', 0, '', '[]'
);

INSERT INTO Host VALUES (
    'host-001', '', 'esxi-host-001.example.com', '{"kind":"Cluster","id":"cluster-1"}', 1,
    'cluster-1', 'green', 0, '192.168.1.100', '', 'UTC',
    2, 16, 137438953472,
    'VMware ESXi', '7.0.3', 'ProLiant DL380 Gen10', 'HPE',
    '{}', '[]',
    '[{"kind":"Datastore","id":"datastore-1"},{"kind":"Datastore","id":"datastore-2"}]',
    '[{"canonicalName":"naa.001","vendor":"ATA","model":"HPE E208i-a SR Gen10","key":"key-001"}]',
    '{}',
    '[{"hbaDevice":"vmhba0","protocol":"SAS","model":"HPE E208i-a SR Gen10","key":"hba-001"}]',
    '[]'
);

INSERT INTO Host VALUES (
    'host-002', '', 'esxi-host-002.example.com', '{"kind":"Cluster","id":"cluster-1"}', 1,
    'cluster-1', 'green', 0, '192.168.1.101', '', 'UTC',
    2, 24, 274877906944,
    'VMware ESXi', '7.0.3', 'PowerEdge R740', 'Dell',
    '{}', '[]',
    '[{"kind":"Datastore","id":"datastore-1"}]',
    '[{"canonicalName":"naa.002","vendor":"NetApp","model":"FibreChannel","key":"key-002"}]',
    '{}',
    '[{"hbaDevice":"vmhba1","protocol":"FibreChannel","model":"QLE2742","key":"hba-002"}]',
    '[]'
);

INSERT INTO Network VALUES (
    'network-1', 'DvPortGroup', 'VM Network', '{"kind":"Folder","id":"folder-1"}', 1,
    '', '{"kind":"DVSwitch","id":"dvs-001"}', 'dvportgroup-1', '[]', '341'
);

INSERT INTO Network VALUES (
    'network-2', 'DvPortGroup', 'Management', '{"kind":"Folder","id":"folder-1"}', 1,
    '', '{"kind":"DVSwitch","id":"dvs-001"}', 'dvportgroup-2', '[]', '200'
);

INSERT INTO Network VALUES (
    'network-3', 'Standard', 'vMotion', '{"kind":"Folder","id":"folder-1"}', 1,
    '', '{}', '', '[]', ''
);

INSERT INTO Datastore VALUES (
    'datastore-1', '', 'datastore1', '{"kind":"Folder","id":"folder-1"}', 1,
    'VMFS', 1099511627776, 536870912000, 'normal', '["naa.001"]'
);

INSERT INTO Datastore VALUES (
    'datastore-2', '', 'datastore2', '{"kind":"Folder","id":"folder-1"}', 1,
    'NFS', 549755813888, 268435456000, 'normal', '[]'
);

INSERT INTO VM VALUES (
    'vm-001', '', 'test-vm-1', '{"kind":"Folder","id":"folder-1"}', 1,
    'folder-1', 'host-001', 1, 1,
    'uuid-001', 'bios', 'poweredOn', 'connected', '[]',
    1, 0, 1, 0,
    4, 2, 8192,
    'Red Hat Enterprise Linux 8 (64-bit)', 'RHEL 8.5',
    'testvm1.example.com', 'rhel8_64Guest', 0, '192.168.1.10',
    '[]', 52428800000, '{}', 0, 1, 0, '[]',
    '[{"network":{"kind":"Network","id":"network-1"},"mac":"00:50:56:aa:bb:01","order":0,"deviceKey":4000}]',
    '[{"key":2000,"unitNumber":0,"controllerKey":1000,"file":"[datastore1] test-vm-1/disk1.vmdk","datastore":{"kind":"Datastore","id":"datastore-1"},"capacity":53687091200,"shared":false,"rdm":false,"bus":"scsi","mode":"persistent","serial":"disk-001","winDriveLetter":"","changeTrackingEnabled":true,"parent":""}]',
    '[]', '[]', '[]', '[]', '[]', '[]', 0, '', '', '', 1, 0
);

INSERT INTO VM VALUES (
    'vm-002', '', 'test-vm-2', '{"kind":"Folder","id":"folder-1"}', 1,
    'folder-1', 'host-001', 1, 1,
    'uuid-002', 'efi', 'poweredOff', 'connected', '[]',
    0, 0, 0, 0,
    2, 2, 4096,
    'Microsoft Windows Server 2019 (64-bit)', 'Windows 2019',
    'testvm2.example.com', 'windows2019srv_64Guest', 512, '192.168.1.11',
    '[]', 31457280000, '{}', 0, 0, 0, '[]',
    '[{"network":{"kind":"Network","id":"network-1"},"mac":"00:50:56:aa:bb:02","order":0,"deviceKey":4000},{"network":{"kind":"Network","id":"network-2"},"mac":"00:50:56:aa:bb:03","order":1,"deviceKey":4001}]',
    '[{"key":2000,"unitNumber":0,"controllerKey":1000,"file":"[datastore1] test-vm-2/disk1.vmdk","datastore":{"kind":"Datastore","id":"datastore-1"},"capacity":53687091200,"shared":false,"rdm":false,"bus":"scsi","mode":"persistent","serial":"disk-002","winDriveLetter":"C:","changeTrackingEnabled":false,"parent":""}]',
    '[]', '[]', '[]', '[]', '[]', '[]', 0, '', '', '', 0, 0
);

INSERT INTO VM VALUES (
    'vm-003', '', 'template-vm', '{"kind":"Folder","id":"folder-1"}', 1,
    'folder-1', 'host-002', 1, 1,
    'uuid-003', 'bios', 'poweredOff', 'connected', '[]',
    0, 0, 0, 0,
    1, 1, 2048,
    'Ubuntu Linux (64-bit)', 'Ubuntu',
    '', 'ubuntu64Guest', 0, '',
    '[]', 10485760000, '{}', 1, 0, 0, '[]',
    '[]',
    '[{"key":2000,"unitNumber":0,"controllerKey":1000,"file":"[datastore2] template-vm/disk1.vmdk","datastore":{"kind":"Datastore","id":"datastore-2"},"capacity":21474836480,"shared":false,"rdm":false,"bus":"scsi","mode":"persistent","serial":"disk-003","winDriveLetter":"","changeTrackingEnabled":false,"parent":""}]',
    '[]', '[]', '[]', '[]', '[]', '[]', 0, '', '', '', 0, 0
);
