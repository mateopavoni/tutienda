package catalog

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ImportRowError describes one CSV data row (1-indexed, header excluded) that failed to import.
type ImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportResult summarizes a CSV import: how many products landed, and which rows didn't.
type ImportResult struct {
	Created int              `json:"created"`
	Errors  []ImportRowError `json:"errors"`
}

// importRow is one successfully-parsed CSV data row, ready to hand to Service.Create.
type importRow struct {
	row     int
	product Product
}

// parseImportCSV reads a CSV (columns: name,category,description,price,currency,imageUrl — matched by
// header name, order doesn't matter) into a list of ready-to-create products plus any row-level parse
// errors (bad price, malformed CSV line). Pure and DB-free so it's unit-testable on its own; ImportCSV
// below is the thin DB-touching wrapper. Every product gets a single "Único"-style variant
// (Create/normalizeVariants auto-generates its SKU from the name) — bulk multi-variant import isn't
// supported, the manual product form already covers that.
func parseImportCSV(r io.Reader) ([]importRow, []ImportRowError, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	cr.FieldsPerRecord = -1 // rows may be shorter than the header (trailing columns omitted)

	header, err := cr.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("read CSV header: %w", err)
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	if _, ok := col["name"]; !ok {
		return nil, nil, fmt.Errorf("%w: CSV must have a 'name' column", ErrInvalidProduct)
	}
	get := func(record []string, key string) string {
		i, ok := col[key]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	var rows []importRow
	var errs []ImportRowError
	row := 0
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		row++
		if err != nil {
			errs = append(errs, ImportRowError{Row: row, Message: err.Error()})
			continue
		}
		name := get(record, "name")
		if name == "" {
			continue // blank row, skip silently
		}
		var priceCents int64
		if raw := get(record, "price"); raw != "" {
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				errs = append(errs, ImportRowError{Row: row, Message: "invalid price: " + raw})
				continue
			}
			priceCents = int64(f*100 + 0.5)
		}
		rows = append(rows, importRow{
			row: row,
			product: Product{
				Name:        name,
				Category:    get(record, "category"),
				Description: get(record, "description"),
				PriceCents:  priceCents,
				Currency:    get(record, "currency"),
				ImageURL:    get(record, "imageUrl"),
				Variants:    []Variant{{}}, // blank row -> Create auto-generates a SKU from the name
			},
		})
	}
	return rows, errs, nil
}

// ImportCSV creates one product per data row of a CSV file. Each row is independent: a bad row (parse
// error or a Create failure, e.g. hitting the plan's product cap) is recorded and the rest of the file
// still imports, so one typo doesn't sink the whole batch. Starting stock isn't set here (catalog has no
// inventory client); the merchant sets it from the product list afterward, same as leaving quantity
// untouched on the manual form.
func (s *Service) ImportCSV(ctx context.Context, tenant string, r io.Reader) (*ImportResult, error) {
	rows, parseErrs, err := parseImportCSV(r)
	if err != nil {
		return nil, err
	}
	result := &ImportResult{Errors: parseErrs}
	for _, row := range rows {
		if _, err := s.Create(ctx, tenant, row.product); err != nil {
			result.Errors = append(result.Errors, ImportRowError{Row: row.row, Message: err.Error()})
			continue
		}
		result.Created++
	}
	return result, nil
}
