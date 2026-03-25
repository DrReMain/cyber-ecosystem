package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type localeItem struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

func main() {
	service := flag.String("service", "", "service name")
	flag.Parse()

	if *service == "" {
		fail("missing --service")
	}

	s, ok := services[*service]
	if !ok {
		fail("unsupported --service=%s", *service)
	}

	reasons, dir := enumValues(s.ReasonName), s.Dir

	for _, l := range languages {
		if err := syncLocaleFile(filepath.Join(dir, "active."+l+".json"), reasons); err != nil {
			fail("sync %s locale failed: %v", l, err)
		}
	}
}

func enumValues(enumName map[int32]string) []string {
	keys := make([]int, 0, len(enumName))
	for k := range enumName {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		values = append(values, enumName[int32(k)])
	}
	return values
}

func syncLocaleFile(path string, reasons []string) error {
	existing := map[string]string{}
	if raw, err := os.ReadFile(path); err == nil {
		var items []localeItem
		if err := json.Unmarshal(raw, &items); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, item := range items {
			existing[item.ID] = item.Translation
		}
	}

	items := make([]localeItem, 0, len(reasons))
	for _, reason := range reasons {
		items = append(items, localeItem{
			ID:          reason,
			Translation: existing[reason],
		})
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func fail(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
