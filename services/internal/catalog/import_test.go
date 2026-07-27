package catalog

import (
	"strings"
	"testing"
)

// TestParseImportCSV pins the CSV import contract: header-name column matching (order doesn't matter),
// price parsed to cents, a blank name skips the row silently, and a bad price is reported without
// aborting the rest of the file.
func TestParseImportCSV(t *testing.T) {
	csv := "category,name,price\n" +
		"Lamps,Concrete Lamp,19.99\n" +
		",,\n" + // blank row, skipped silently
		"Lamps,Bad Price Lamp,not-a-number\n"

	rows, errs, err := parseImportCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1: %+v", len(rows), rows)
	}
	if rows[0].product.Name != "Concrete Lamp" || rows[0].product.Category != "Lamps" {
		t.Errorf("row not matched by header name: %+v", rows[0].product)
	}
	if rows[0].product.PriceCents != 1999 {
		t.Errorf("PriceCents = %d, want 1999", rows[0].product.PriceCents)
	}
	if len(rows[0].product.Variants) != 1 {
		t.Errorf("want a single blank variant so Create auto-generates a SKU, got %v", rows[0].product.Variants)
	}
	if len(errs) != 1 || errs[0].Row != 3 {
		t.Errorf("errs = %+v, want one error on row 3 (bad price)", errs)
	}
}

func TestParseImportCSVRequiresNameColumn(t *testing.T) {
	_, _, err := parseImportCSV(strings.NewReader("category,price\nLamps,10\n"))
	if err == nil {
		t.Fatal("want an error when the CSV has no 'name' column")
	}
}
