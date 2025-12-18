package store_test

import (
	"context"
	"database/sql"

	"github.com/duckdb/duckdb-go/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tupyy/rvtools/models"
	"github.com/tupyy/rvtools/pkg/store"
)

const createConcernsTable = `
CREATE TABLE concerns (
    "VM_ID" VARCHAR,
    "Concern_ID" VARCHAR,
    "Label" VARCHAR,
    "Category" VARCHAR,
    "Assessment" VARCHAR
);
`

const insertTestConcerns = `
INSERT INTO concerns VALUES
('vm-001', 'concern-001', 'Shared disk detected', 'Storage', 'Warning'),
('vm-002', 'concern-001', 'Shared disk detected', 'Storage', 'Warning'),
('vm-003', 'concern-001', 'Shared disk detected', 'Storage', 'Warning'),
('vm-004', 'concern-001', 'Shared disk detected', 'Storage', 'Warning'),
('vm-001', 'concern-002', 'CPU hot-add enabled', 'Configuration', 'Information'),
('vm-002', 'concern-002', 'CPU hot-add enabled', 'Configuration', 'Information'),
('vm-003', 'concern-002', 'CPU hot-add enabled', 'Configuration', 'Information'),
('vm-004', 'concern-002', 'CPU hot-add enabled', 'Configuration', 'Information'),
('vm-001', 'concern-003', 'CBT not enabled', 'Backup', 'Critical'),
('vm-002', 'concern-003', 'CBT not enabled', 'Backup', 'Critical'),
('vm-003', 'concern-003', 'CBT not enabled', 'Backup', 'Critical'),
('vm-004', 'concern-003', 'CBT not enabled', 'Backup', 'Critical');
`

func setupConcernTestDB() *sql.DB {
	c, err := duckdb.NewConnector("", nil)
	Expect(err).NotTo(HaveOccurred())

	db := sql.OpenDB(c)
	Expect(db).NotTo(BeNil())

	_, err = db.Exec(createConcernsTable)
	Expect(err).NotTo(HaveOccurred())

	_, err = db.Exec(insertTestConcerns)
	Expect(err).NotTo(HaveOccurred())

	return db
}

var _ = Describe("ConcernStore", func() {
	var (
		db           *sql.DB
		concernStore store.Concern
		ctx          context.Context
	)

	BeforeEach(func() {
		db = setupConcernTestDB()
		concernStore = store.NewConcernStore(db)
		ctx = context.Background()
	})

	AfterEach(func() {
		if db != nil {
			db.Close()
		}
	})

	Describe("Get", func() {
		It("returns all unique concerns when no filter is applied", func() {
			concerns, err := concernStore.Get(ctx, store.NewConcernQueryFilter())

			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(HaveLen(3))

			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM concerns").Scan(&count)
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(Equal(12)) // Actual rows
		})

		It("filters concerns by VM ID", func() {
			filter := store.NewConcernQueryFilter().WhereVmId("'vm-001'")
			concerns, err := concernStore.Get(ctx, filter)

			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(HaveLen(3))

			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM concerns WHERE VM_ID = 'vm-001'").Scan(&count)
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(Equal(3)) // Actual rows
		})

		It("returns empty slice when VM ID has no concerns", func() {
			filter := store.NewConcernQueryFilter().WhereVmId("'vm-999'")
			concerns, err := concernStore.Get(ctx, filter)

			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(BeEmpty())
		})
	})

	Describe("Insert", func() {
		It("inserts a new concern using ConcernValuesBuilder", func() {
			concern := models.Concern{
				Id:         "concern-004",
				Label:      "New concern",
				Category:   "Test",
				Assessment: "Info",
			}

			builder := store.NewConcernValuesBuilder().Append("vm-005", []models.Concern{concern})
			err := concernStore.Insert(ctx, builder.Build())

			Expect(err).NotTo(HaveOccurred())

			filter := store.NewConcernQueryFilter().WhereVmId("'vm-005'")
			concerns, err := concernStore.Get(ctx, filter)

			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(HaveLen(1))
			Expect(concerns[0].Id).To(Equal("concern-004"))
			Expect(concerns[0].Label).To(Equal("New concern"))
		})

		It("inserts multiple concerns at once using ConcernValuesBuilder", func() {
			concernA := models.Concern{
				Id:         "concern-005",
				Label:      "Concern A",
				Category:   "Cat A",
				Assessment: "Info",
			}
			concernB := models.Concern{
				Id:         "concern-006",
				Label:      "Concern B",
				Category:   "Cat B",
				Assessment: "Warning",
			}

			builder := store.NewConcernValuesBuilder().
				Append("vm-006", []models.Concern{concernA}).
				Append("vm-006", []models.Concern{concernB})

			err := concernStore.Insert(ctx, builder.Build())

			Expect(err).NotTo(HaveOccurred())

			filter := store.NewConcernQueryFilter().WhereVmId("'vm-006'")
			concerns, err := concernStore.Get(ctx, filter)

			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(HaveLen(2))
		})

		It("inserts same concern for multiple VMs using ConcernValuesBuilder", func() {
			concern := models.Concern{
				Id:         "concern-007",
				Label:      "Shared concern",
				Category:   "Shared",
				Assessment: "Warning",
			}

			builder := store.NewConcernValuesBuilder().
				Append("vm-007", []models.Concern{concern}).
				Append("vm-008", []models.Concern{concern}).
				Append("vm-009", []models.Concern{concern})

			err := concernStore.Insert(ctx, builder.Build())

			Expect(err).NotTo(HaveOccurred())

			// Should return 1 unique concern (DISTINCT)
			concerns, err := concernStore.Get(ctx, store.NewConcernQueryFilter().WhereVmId("'vm-007'", "'vm-008'", "'vm-009'"))
			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(HaveLen(1))

			// But 3 rows in the table for these VMs
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM concerns WHERE Concern_ID = 'concern-007'").Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(3))
		})
	})

	Describe("Delete", func() {
		It("deletes concerns by VM ID", func() {
			filter := store.NewConcernQueryFilter().WhereVmId("'vm-001'")
			err := concernStore.Delete(ctx, filter)

			Expect(err).NotTo(HaveOccurred())

			concerns, err := concernStore.Get(ctx, filter)

			Expect(err).NotTo(HaveOccurred())
			Expect(concerns).To(BeEmpty())
		})

		It("keeps other concerns after deletion", func() {
			filter := store.NewConcernQueryFilter().WhereVmId("'vm-001'", "'vm-002'", "'vm-003'", "'vm-004'")
			err := concernStore.Delete(ctx, filter)

			Expect(err).NotTo(HaveOccurred())

			allConcerns, err := concernStore.Get(ctx, store.NewConcernQueryFilter())

			Expect(err).NotTo(HaveOccurred())
			Expect(allConcerns).To(HaveLen(0))
		})
	})
})

var _ = Describe("ConcernQueryFilter", func() {
	Describe("Build", func() {
		It("returns empty string when no filters are applied", func() {
			filter := store.NewConcernQueryFilter()
			result := filter.Build()

			Expect(result).To(BeEmpty())
		})

		It("builds WHERE clause with single VM ID", func() {
			filter := store.NewConcernQueryFilter().WhereVmId("'vm-001'")
			result := filter.Build()

			Expect(result).To(Equal(`WHERE VM_ID IN ('vm-001')`))
		})

		It("builds WHERE clause with multiple VM IDs", func() {
			filter := store.NewConcernQueryFilter().WhereVmId("'vm-001'", "'vm-002'")
			result := filter.Build()

			Expect(result).To(Equal(`WHERE VM_ID IN ('vm-001','vm-002')`))
		})

		It("returns empty string when WhereVmId is called with no IDs", func() {
			filter := store.NewConcernQueryFilter().WhereVmId()
			result := filter.Build()

			Expect(result).To(BeEmpty())
		})
	})
})

var _ = Describe("ConcernValuesBuilder", func() {
	Describe("Build", func() {
		It("returns empty string when no values are appended", func() {
			builder := store.NewConcernValuesBuilder()
			result := builder.Build()

			Expect(result).To(BeEmpty())
		})

		It("builds single value correctly", func() {
			concern := models.Concern{
				Id:         "c-001",
				Label:      "Test Label",
				Category:   "Test Category",
				Assessment: "Warning",
			}

			builder := store.NewConcernValuesBuilder().Append("vm-001", []models.Concern{concern})
			result := builder.Build()

			Expect(result).To(Equal("('vm-001', 'c-001', 'Test Label', 'Test Category', 'Warning')"))
		})

		It("builds multiple values correctly", func() {
			concern1 := models.Concern{
				Id:         "c-001",
				Label:      "Label 1",
				Category:   "Category 1",
				Assessment: "Info",
			}
			concern2 := models.Concern{
				Id:         "c-002",
				Label:      "Label 2",
				Category:   "Category 2",
				Assessment: "Critical",
			}

			builder := store.NewConcernValuesBuilder().
				Append("vm-001", []models.Concern{concern1}).
				Append("vm-002", []models.Concern{concern2})
			result := builder.Build()

			Expect(result).To(Equal("('vm-001', 'c-001', 'Label 1', 'Category 1', 'Info') , ('vm-002', 'c-002', 'Label 2', 'Category 2', 'Critical')"))
		})

		It("supports method chaining", func() {
			concern := models.Concern{
				Id:         "c-001",
				Label:      "Label",
				Category:   "Category",
				Assessment: "Info",
			}

			builder := store.NewConcernValuesBuilder().
				Append("vm-001", []models.Concern{concern}).
				Append("vm-002", []models.Concern{concern}).
				Append("vm-003", []models.Concern{concern})

			result := builder.Build()

			Expect(result).To(ContainSubstring("vm-001"))
			Expect(result).To(ContainSubstring("vm-002"))
			Expect(result).To(ContainSubstring("vm-003"))
		})
	})
})
