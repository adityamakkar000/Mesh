package build

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

func writeToTar(path string, tw *tar.Writer, fi os.FileInfo) {
	fr, _ := os.Open(path)
	defer fr.Close()

	h := new(tar.Header)
	if fi.IsDir() {
		h.Typeflag = tar.TypeDir
	} else {
		h.Typeflag = tar.TypeReg
	}
	root, _ := filepath.Abs("./")
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	h.Name = rel
	h.Size = fi.Size()
	h.Mode = int64(fi.Mode())
	h.ModTime = fi.ModTime()
	_ = tw.WriteHeader(h)

	if !fi.IsDir() {
		_, _ = io.Copy(tw, fr)
	}
}

func getFilesToSend(ignore []string) []string {

	var files []string

	root, err := filepath.Abs("./")
	if err != nil {
		ui.ErrorWrap(err, "failed to get absolute path of current directory")
		return files
	}

	errWalk := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if shouldIgnore(path, ignore) {
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

func makeTarStream(files []string, w io.Writer) {
	tw := tar.NewWriter(w)
	defer tw.Close()

	for _, file := range files {

		fi, err := os.Stat(file)
		if err != nil {
			continue
		}
		writeToTar(file, tw, fi)
	}
	ui.Info("Wrote files to tar stream")
}

func BuildTar() {
	var meshConfig, err = parse.Mesh()
	if err != nil {
		ui.Error("Could not parse mesh.yaml")
	}
	var ignore = meshConfig.Ignore
	files := getFilesToSend(ignore)
	out, err := os.Create("mesh.tar")
	if err != nil {
		ui.Error("Could not make mesh.tar")
	}

	defer out.Close()
	makeTarStream(files, out)


}

func sendTar(){

}
