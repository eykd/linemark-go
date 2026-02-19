package cmd

import (
	"encoding/json"
	"fmt"
	"io"
)

// writeJSON encodes v as JSON to w, handling I/O errors at the boundary.
func writeJSON(w io.Writer, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintf(w, "{\"error\":%q}\n", err.Error())
	}
}
