// Command clean-gen wipes a single gen subtree's derived files before buf
// generate (buf is purely additive and never removes files whose source proto
// was deleted). Hand-maintained package files at the target root — and
// node_modules anywhere — survive; everything else goes, and directories that
// become empty collapse bottom-up. Enumeration and deletion both go through
// os.Root, which confines them to the target subtree by construction. A
// missing target is a no-op (fresh checkout with no gen/, or a wiped tree).
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
)

// handMaintained are root-level files buf/hey-api do NOT produce and that a
// regen must not wipe: they declare the workspace package (package.json,
// tsconfig.json) or the generator config co-located in its package
// (hey-api.config.ts). Matched only at the target root, never inside
// generated sub-packages.
var handMaintained = map[string]bool{
	"package.json":      true,
	"tsconfig.json":     true,
	"hey-api.config.ts": true,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "clean-gen: missing target directory argument (e.g. gen/go)")
		os.Exit(2)
	}
	root, err := os.OpenRoot(os.Args[1])
	if os.IsNotExist(err) {
		return // fresh checkout / manually wiped tree
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "clean-gen:", err)
		os.Exit(1)
	}
	if err := clean(root, ".", true); err != nil {
		_ = root.Close()
		fmt.Fprintln(os.Stderr, "clean-gen:", err)
		os.Exit(1)
	}
	if err := root.Close(); err != nil {
		fmt.Fprintln(os.Stderr, "clean-gen: close:", err)
		os.Exit(1)
	}
}

func clean(root *os.Root, dir string, isRoot bool) error {
	fsys := root.FS()
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := path.Join(dir, e.Name())
		if isRoot && e.Type().IsRegular() && handMaintained[e.Name()] {
			continue
		}
		// node_modules is pnpm install state, not generated output: wiping it
		// races `go mod tidy`/`hey-api` against pnpm's rebuild window and
		// forces a full reinstall on every regeneration.
		if e.IsDir() && e.Name() == "node_modules" {
			continue
		}
		if e.IsDir() {
			if err := clean(root, name, false); err != nil {
				return err
			}
			// Collapse subdirectories that became empty after wiping;
			// Remove declines non-empty ones (a node_modules survivor).
			leftover, lerr := fs.ReadDir(fsys, name)
			if lerr != nil {
				return lerr
			}
			if len(leftover) == 0 {
				if err := root.Remove(name); err != nil {
					return err
				}
			}
			continue
		}
		if err := root.Remove(name); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
