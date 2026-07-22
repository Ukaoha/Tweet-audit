package report

import (
	"encoding/csv"
	"fmt"
	"os"

	"tweet-audit/audit"
)

// WriteCSV writes verdicts to a CSV file suitable for Twitter deletion.
func WriteCSV(path string, verdicts []audit.Verdict) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating csv file: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)

	// Header
	w.Write([]string{"Tweet URL", "Tweet Text", "Created At", "Flag", "Reason"})

	// Rows
	for _, v := range verdicts {
		w.Write([]string{v.URL, v.Text, v.CreatedAt, fmt.Sprintf("%v", v.Flag), v.Reason})
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flushing csv: %w", err)
	}

	return nil
}
