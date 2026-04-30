package prerun

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adityamakkar000/Mesh/internal/parse"
	"github.com/adityamakkar000/Mesh/internal/ui"
)

func shouldIgnore(path string, ignore []string) bool {
	for _, pattern := range ignore {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func getFilesToSend(copyRoot string, ignore []string) []string {

	var files []string

	errWalk := filepath.WalkDir(copyRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(copyRoot, path)
		if shouldIgnore(relativePath, ignore) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if errWalk != nil {
		ui.ErrorWrap(errWalk, "failed to walk directory for copying files")
		return files
	}

	return files
}

func writeToTar(copyRoot, path string, tw *tar.Writer, fi os.FileInfo) error {
	fr, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fr.Close()

	rel, _ := filepath.Rel(copyRoot, path)

	h, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	h.Name = rel

	if err := tw.WriteHeader(h); err != nil {
		return err
	}

	if !fi.IsDir() {
		_, err = io.Copy(tw, fr)
		return err
	}
	return nil
}

func BuildTar(meshYAMLDir string) io.Reader {
	meshConfig, err := parse.Mesh(meshYAMLDir)
	if err != nil {
		ui.Error("Could not parse mesh.yaml")
		return nil
	}

	copyRoot, err := filepath.Abs(".")
	if err != nil {
		ui.ErrorWrap(err, "failed to get absolute path of current directory for upload")
		return nil
	}

	files := getFilesToSend(copyRoot, meshConfig.Ignore)

	r, w := io.Pipe()

	go func() {
		defer w.Close()
		tw := tar.NewWriter(w)
		defer tw.Close()
		for _, file := range files {
			fi, err := os.Stat(file)
			if err != nil {
				_ = w.CloseWithError(err)
				return
			}
			if err := writeToTar(copyRoot, file, tw, fi); err != nil {
				_ = w.CloseWithError(err)
				return
			}
		}
	}()

	return r
}
