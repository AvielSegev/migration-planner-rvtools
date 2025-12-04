package definitions

import "fmt"

type Datastore struct {
	FreeCapacity  float32 `db:"freeCapacity" json:"freeCapacity"`
	Mha           string  `db:"mha" json:"mha"`
	Hosts         string  `db:"hosts" json:"hosts"`
	DiskType      string  `db:"diskType" json:"diskType"`
	TotalCapacity string  `db:"totalCapacity" json:"totalCapacity"`
	HbaModel      string  `db:"hbaModel" json:"hbaModel"`
	HbaType       string  `db:"hbaType" json:"hbaType"`
}

func (d Datastore) String() string {
	return fmt.Sprintf("Datastore{Capacity: %s (%.2f%% free), DiskType: %s, HBA: %s/%s, Hosts: %s, MHA: %s}",
		d.TotalCapacity, d.FreeCapacity, d.DiskType, d.HbaType, d.HbaModel, d.Hosts, d.Mha)
}

type Os struct {
	Name  string `db:"name" json:"name"`
	Count int    `db:"count" json:"count"`
}
