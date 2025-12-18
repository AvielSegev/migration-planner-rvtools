package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/georgysavva/scany/v2/sqlscan"
	"strings"

	"github.com/tupyy/rvtools/models"
)

const (
	VMIDCol       = "VM_ID"
	ConcernIDCol  = "Concern_ID"
	LabelCol      = "Label"
	CategoryCol   = "Category"
	AssessmentCol = "Assessment"
)

const (
	insertConcernStm = "INSERT INTO concerns (%s, %s, %s, %s, %s) VALUES %s;"
	DeleteConcernStm = "DELETE FROM concerns %s;"
	SelectConcernStm = "SELECT DISTINCT %s, %s, %s, %s FROM concerns %s;"
)

type Concern interface {
	Get(ctx context.Context, filter *ConcernQueryFilter) ([]models.Concern, error)
	Insert(ctx context.Context, values string) error
	Delete(ctx context.Context, filter *ConcernQueryFilter) error
}

func NewConcernStore(db *sql.DB) Concern {
	return &ConcernStore{db: db}
}

type ConcernStore struct {
	db *sql.DB
}

func (c ConcernStore) Get(ctx context.Context, filter *ConcernQueryFilter) ([]models.Concern, error) {
	query := fmt.Sprintf(
		SelectConcernStm,
		ConcernIDCol, LabelCol, CategoryCol, AssessmentCol,
		filter.Build(),
	)

	var concerns []models.Concern
	if err := sqlscan.Select(ctx, c.db, &concerns, query); err != nil {
		return nil, fmt.Errorf("scanning concerns: %w", err)
	}

	return concerns, nil
}

func (c ConcernStore) Insert(ctx context.Context, values string) error {
	// values must be:
	// ('vm1','cid1','lbl','cat1','asm1'), ('vm2','cid2','lbl2','cat2','asm2')
	query := fmt.Sprintf(insertConcernStm, VMIDCol, ConcernIDCol, LabelCol, CategoryCol, AssessmentCol, values)

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("insert concerns failed: %w", err)
	}

	return nil
}

func (c ConcernStore) Delete(ctx context.Context, filter *ConcernQueryFilter) error {
	query := fmt.Sprintf(DeleteConcernStm, filter.Build())

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("delete concerns failed: %w", err)
	}

	return nil
}

type ConcernQueryFilter struct {
	filters map[string]string
}

func NewConcernQueryFilter() *ConcernQueryFilter {
	return &ConcernQueryFilter{make(map[string]string)}
}

func (cf *ConcernQueryFilter) WhereVmId(Ids ...string) *ConcernQueryFilter {
	if len(Ids) == 0 {
		return cf
	}

	idFilterKey := fmt.Sprintf("%s IN", VMIDCol)
	idFilterVal := fmt.Sprintf("(%s)", strings.Join(Ids, ","))

	cf.filters[idFilterKey] = idFilterVal
	return cf
}

func (cf *ConcernQueryFilter) Build() string {
	if len(cf.filters) == 0 {
		return ""
	}

	var filters []string

	for k, v := range cf.filters {
		filters = append(filters, fmt.Sprintf("%s %s", k, v))
	}

	return fmt.Sprintf("WHERE %s", strings.Join(filters, " AND "))
}

type ConcernValuesBuilder struct {
	values []string
}

func NewConcernValuesBuilder() *ConcernValuesBuilder {
	return &ConcernValuesBuilder{}
}

func (cb *ConcernValuesBuilder) Append(vmId string, concerns []models.Concern) *ConcernValuesBuilder {
	escape := func(s string) string {
		return strings.ReplaceAll(s, "'", "''")
	}

	for _, c := range concerns {
		value := fmt.Sprintf("('%s', '%s', '%s', '%s', '%s')",
			escape(vmId),
			escape(c.Id),
			escape(c.Label),
			escape(c.Category),
			escape(c.Assessment),
		)
		cb.values = append(cb.values, value)
	}
	return cb
}

func (cb *ConcernValuesBuilder) Build() string {
	if len(cb.values) == 0 {
		return ""
	}

	return strings.Join(cb.values, ", ")
}
