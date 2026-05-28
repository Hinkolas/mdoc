// Package bundle packages a document and its dependencies into a portable
// .mdoc zip archive. The bundle is a regular zip with a custom extension —
// same pattern as .docx / .epub / .pptx — so any unzip tool can crack it
// open and the file manager can preview the contents.
package bundle

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hinkolas/mdoc/internal/document"
	"github.com/hinkolas/mdoc/internal/theme"
)

// Options configures Export.
type Options struct {
	// OutputPath is the .mdoc file to write. Empty means write next to
	// the source document as <basename>.mdoc.
	OutputPath string
}

// Result describes what Export produced.
type Result struct {
	OutputPath string
	// Entries lists the bundle-relative paths included, in insertion order.
	Entries []string
}

// Export packs the document, its resolved theme, and the conventional
// assets/ sibling directory (if present) into a .mdoc zip. The bundle is
// laid out so that unzipping it yields a directory mdoc can open
// directly:
//
//	example.mdoc
//	├── document.md
//	├── themes/<name>.html
//	└── assets/...
func Export(doc *document.Document, thm *theme.Theme, opts Options) (*Result, error) {
	outPath := opts.OutputPath
	if outPath == "" {
		base := filepath.Base(doc.Path)
		ext := filepath.Ext(base)
		outPath = filepath.Join(filepath.Dir(doc.Path), base[:len(base)-len(ext)]+".mdoc")
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return nil, fmt.Errorf("resolve output path: %w", err)
	}

	f, err := os.Create(absOut)
	if err != nil {
		return nil, fmt.Errorf("create bundle: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	res := &Result{OutputPath: absOut}

	// 1. The source document, at the bundle root with its original name
	//    so the unpacked layout works as a normal mdoc project.
	docEntry := filepath.Base(doc.Path)
	if err := addFile(zw, docEntry, doc.Path); err != nil {
		return nil, fmt.Errorf("add document: %w", err)
	}
	res.Entries = append(res.Entries, docEntry)

	// 2. The resolved theme. Always included regardless of whether it
	//    came from the project's themes/ or the user's config dir — a
	//    bundle should be self-contained.
	if thm.Path != "" {
		themeEntry := filepath.ToSlash(filepath.Join("themes", thm.Name+".html"))
		if err := addFile(zw, themeEntry, thm.Path); err != nil {
			return nil, fmt.Errorf("add theme: %w", err)
		}
		res.Entries = append(res.Entries, themeEntry)
	}

	// 3. Optional assets/ sibling. Mirrored verbatim so relative
	//    references inside the document keep resolving after unpack.
	assetsDir := filepath.Join(doc.Dir, "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		walkErr := filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(doc.Dir, path)
			if err != nil {
				return err
			}
			entry := filepath.ToSlash(rel)
			if err := addFile(zw, entry, path); err != nil {
				return err
			}
			res.Entries = append(res.Entries, entry)
			return nil
		})
		if walkErr != nil {
			return nil, fmt.Errorf("add assets: %w", walkErr)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finalize bundle: %w", err)
	}
	return res, nil
}

// addFile copies sourcePath into the zip at bundlePath, preserving mtime
// and a basic mode. Names are normalized to forward slashes per the zip
// spec so the bundle is portable across operating systems.
func addFile(zw *zip.Writer, bundlePath, sourcePath string) error {
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(bundlePath)
	header.Method = zip.Deflate

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, src)
	return err
}
