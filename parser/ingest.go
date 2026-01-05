package parser

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

var stmtRegex = regexp.MustCompile(`(?s)(CREATE|INSERT|UPDATE|DROP|ALTER|WITH|INSTALL|LOAD|ATTACH|DETACH).*?;`)

// IngestRvTools ingests data from an RVTools Excel file and runs validation if a validator is configured.
// TODO: Add file validation (empty/truncated detection) and schema validation (required sheets/columns)
func (p *Parser) IngestRvTools(ctx context.Context, excelFile string) error {
	query, err := p.builder.IngestRvtoolsQuery(excelFile)
	if err != nil {
		return fmt.Errorf("building rvtools ingestion query: %w", err)
	}
	if err := p.executeStatements(query); err != nil {
		return fmt.Errorf("ingesting rvtools data: %w", err)
	}
	return p.validate(ctx)
}

// IngestSqlite ingests data from a forklift SQLite database and runs validation if a validator is configured.
func (p *Parser) IngestSqlite(ctx context.Context, sqliteFile string) error {
	query, err := p.builder.IngestSqliteQuery(sqliteFile)
	if err != nil {
		return fmt.Errorf("building sqlite ingestion query: %w", err)
	}
	if err := p.executeStatements(query); err != nil {
		return fmt.Errorf("ingesting sqlite data: %w", err)
	}
	return p.validate(ctx)
}

// executeStatements executes a multi-statement SQL string.
// Errors are logged but not returned since missing sheets are expected in RVTools exports.
func (p *Parser) executeStatements(query string) error {
	stmts := stmtRegex.FindAllString(query, -1)
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := p.db.Exec(stmt); err != nil {
			zap.S().Debugw("statement failed", "error", err)
		}
	}
	return nil
}

// validate is the internal implementation of validation.
func (p *Parser) validate(ctx context.Context) error {
	if p.validator == nil {
		return nil
	}

	vms, err := p.VMs(ctx, Filters{}, Options{})
	if err != nil {
		return fmt.Errorf("getting VMs for validation: %w", err)
	}

	builder := NewConcernValuesBuilder()
	for _, vm := range vms {
		concerns, err := p.validator.Validate(ctx, vm)
		if err != nil {
			zap.S().Warnw("validation failed for VM", "vm_id", vm.ID, "error", err)
			continue
		}
		builder.Append(vm.ID, concerns...)
	}

	return InsertConcerns(ctx, p.db, builder)
}
