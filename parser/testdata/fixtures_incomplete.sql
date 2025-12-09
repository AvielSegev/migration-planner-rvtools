-- Test fixtures for incomplete RVTools exports (like dallas.xlsx)
-- Missing sheets: dvport, vhba

-- vinfo: VM information
CREATE TABLE vinfo (
    "VM ID" VARCHAR,
    "VM" VARCHAR,
    "Folder ID" VARCHAR,
    "Host" VARCHAR,
    "SMBIOS UUID" VARCHAR,
    "Firmware" VARCHAR,
    "Powerstate" VARCHAR,
    "Connection state" VARCHAR,
    "FT State" VARCHAR,
    "CPUs" INTEGER,
    "Memory" INTEGER,
    "OS according to the configuration file" VARCHAR,
    "OS according to the VMware Tools" VARCHAR,
    "DNS Name" VARCHAR,
    "Primary IP Address" VARCHAR,
    "In Use MiB" INTEGER,
    "Template" VARCHAR,
    "CBT" VARCHAR,
    "EnableUUID" VARCHAR,
    "Datacenter" VARCHAR,
    "Cluster" VARCHAR,
    "HW version" VARCHAR,
    "Total disk capacity MiB" INTEGER,
    "Provisioned MiB" INTEGER,
    "Resource pool" VARCHAR,
    "VI SDK UUID" VARCHAR,
    "Network #1" VARCHAR,
    "Network #2" VARCHAR
);

INSERT INTO vinfo VALUES
('vm-001', 'test-vm-1', 'folder-1', 'host-001', 'uuid-001', 'bios', 'poweredOn', 'connected', 'Not protected', 4, 8192, 'Red Hat Enterprise Linux 8', 'RHEL 8.5', 'testvm1.example.com', '192.168.1.10', 50000, 'False', 'True', 'True', 'TestDC', 'TestCluster', 'vmx-19', 102400, 204800, 'Resources', 'vcenter-uuid-001', 'network-001', 'network-002'),
('vm-002', 'test-vm-2', 'folder-1', 'host-001', 'uuid-002', 'efi', 'poweredOff', 'connected', 'Not protected', 2, 4096, 'Microsoft Windows Server 2019', 'Windows 2019', 'testvm2.example.com', '192.168.1.11', 30000, 'False', 'False', 'False', 'TestDC', 'TestCluster', 'vmx-17', 51200, 102400, 'Resources', 'vcenter-uuid-001', 'network-001', NULL);

-- vcpu: CPU details
CREATE TABLE vcpu (
    "VM ID" VARCHAR,
    "VM" VARCHAR,
    "Host" VARCHAR,
    "Cluster" VARCHAR,
    "Datacenter" VARCHAR,
    "Sockets" INTEGER,
    "Cores p/s" INTEGER,
    "CPUs" INTEGER,
    "Hot Add" VARCHAR,
    "Hot Remove" VARCHAR
);

INSERT INTO vcpu VALUES
('vm-001', 'test-vm-1', 'host-001', 'TestCluster', 'TestDC', 2, 2, 4, 'True', 'False'),
('vm-002', 'test-vm-2', 'host-001', 'TestCluster', 'TestDC', 1, 2, 2, 'False', 'False');

-- vmemory: Memory details
CREATE TABLE vmemory (
    "VM ID" VARCHAR,
    "VM" VARCHAR,
    "Host" VARCHAR,
    "Cluster" VARCHAR,
    "Datacenter" VARCHAR,
    "Memory" INTEGER,
    "Hot Add" VARCHAR,
    "Ballooned" INTEGER
);

INSERT INTO vmemory VALUES
('vm-001', 'test-vm-1', 'host-001', 'TestCluster', 'TestDC', 8192, 'True', 0),
('vm-002', 'test-vm-2', 'host-001', 'TestCluster', 'TestDC', 4096, 'False', 512);

-- vdisk: Disk details
CREATE TABLE vdisk (
    "VM ID" VARCHAR,
    "VM" VARCHAR,
    "Host" VARCHAR,
    "Cluster" VARCHAR,
    "Datacenter" VARCHAR,
    "Disk Key" VARCHAR,
    "Unit #" VARCHAR,
    "Path" VARCHAR,
    "Capacity MiB" INTEGER,
    "Sharing mode" VARCHAR,
    "Raw" VARCHAR,
    "Shared Bus" VARCHAR,
    "Disk Mode" VARCHAR,
    "Disk UUID" VARCHAR,
    "Thin" VARCHAR,
    "Controller" VARCHAR,
    "Label" VARCHAR,
    "SCSI Unit #" VARCHAR
);

INSERT INTO vdisk VALUES
('vm-001', 'test-vm-1', 'host-001', 'TestCluster', 'TestDC', '2000', '0', '[datastore1] test-vm-1/disk1.vmdk', 51200, 'sharingNone', 'False', 'scsi', 'persistent', 'disk-uuid-001', 'True', 'SCSI controller 0', 'Hard disk 1', '0'),
('vm-002', 'test-vm-2', 'host-001', 'TestCluster', 'TestDC', '2000', '0', '[datastore2] test-vm-2/disk1.vmdk', 51200, 'sharingNone', 'False', 'scsi', 'persistent', 'disk-uuid-003', 'True', 'SCSI controller 0', 'Hard disk 1', '0');

-- vnetwork: Network interface details
CREATE TABLE vnetwork (
    "VM ID" VARCHAR,
    "VM" VARCHAR,
    "Host" VARCHAR,
    "Cluster" VARCHAR,
    "Datacenter" VARCHAR,
    "Network" VARCHAR,
    "Mac Address" VARCHAR,
    "NIC label" VARCHAR,
    "Adapter" VARCHAR,
    "Switch" VARCHAR,
    "Connected" VARCHAR,
    "Starts Connected" VARCHAR,
    "Type" VARCHAR,
    "IPv4 Address" VARCHAR,
    "IPv6 Address" VARCHAR
);

INSERT INTO vnetwork VALUES
('vm-001', 'test-vm-1', 'host-001', 'TestCluster', 'TestDC', 'VM Network', '00:50:56:aa:bb:01', 'Network adapter 1', 'vmxnet3', 'dvs-001', 'True', 'True', 'distributed', '192.168.1.10', ''),
('vm-002', 'test-vm-2', 'host-001', 'TestCluster', 'TestDC', 'VM Network', '00:50:56:aa:bb:03', 'Network adapter 1', 'e1000', '', 'True', 'True', 'standard', '192.168.1.11', '');

-- vhost: Host details
CREATE TABLE vhost (
    "Host" VARCHAR,
    "Cluster" VARCHAR,
    "Datacenter" VARCHAR,
    "Object ID" VARCHAR,
    "# CPU" INTEGER,
    "# Cores" INTEGER,
    "# Memory" INTEGER,
    "Model" VARCHAR,
    "Vendor" VARCHAR
);

INSERT INTO vhost VALUES
('host-001', 'TestCluster', 'TestDC', 'host-001', 2, 16, 131072, 'ProLiant DL380 Gen10', 'HPE');

-- vdatastore: Datastore details
CREATE TABLE vdatastore (
    "Name" VARCHAR,
    "Address" VARCHAR,
    "Hosts" VARCHAR,
    "Free MiB" INTEGER,
    "Capacity MiB" INTEGER,
    "MHA" VARCHAR,
    "Type" VARCHAR,
    "Datacenter" VARCHAR
);

INSERT INTO vdatastore VALUES
('datastore1', 'naa.001', 'host-001', 512000, 1048576, 'True', 'VMFS', 'TestDC');

-- NOTE: dvport and vhba tables are intentionally missing to simulate incomplete exports
