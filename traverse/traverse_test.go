package traverse

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunnerWalk(t *testing.T) {
	t.Run("No files change should return last modify time", func(t *testing.T) {
		now := time.Now()
		dir := "."

		mod := IsModify(dir, now, nil)

		assert.False(t, mod, "should return last modify time.")
	})

	t.Run("should return trure when file has change", func(t *testing.T) {
		form := "Mon Jan _2 15:04:05 2006"
		lastMod, _ := time.Parse(form, "Sat Feb 08 07:00:00 1992")
		dir := "."

		mod := IsModify(dir, lastMod, nil)

		assert.True(t, mod, "should return lastest modify time.")
	})
}

// info implements fs.DirEntry for testing.
type info struct{}

func (i info) Name() string               { return "" }
func (i info) IsDir() bool                { return true }
func (i info) Type() fs.FileMode          { return fs.ModeDir }
func (i info) Info() (fs.FileInfo, error) { return fakeFileInfo{}, nil }

// fakeFileInfo implements fs.FileInfo with zero ModTime.
type fakeFileInfo struct{}

func (fakeFileInfo) Name() string      { return "" }
func (fakeFileInfo) Size() int64       { return 0 }
func (fakeFileInfo) Mode() fs.FileMode { return fs.ModeDir }
func (fakeFileInfo) ModTime() time.Time { return time.Time{} } // zero — never "after" a recent lastMod
func (fakeFileInfo) IsDir() bool       { return true }
func (fakeFileInfo) Sys() any          { return nil }

// errCloseRC is an io.ReadCloser that reads from a string and errors on Close().
type errCloseRC struct {
	r io.Reader
}

func (e *errCloseRC) Read(p []byte) (int, error) { return e.r.Read(p) }
func (e *errCloseRC) Close() error               { return errors.New("mock close error") }

func TestReadGitignore(t *testing.T) {
	t.Run("no .gitignore file returns nil", func(t *testing.T) {
		dir := t.TempDir()
		patterns := ReadGitignore(dir)
		assert.Nil(t, patterns)
	})

	t.Run("parses comments, empty lines, patterns, and slash-prefixed lines", func(t *testing.T) {
		dir := t.TempDir()
		content := "# comment\n\n*.log\n/vendor\nbuild\n"
		err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644)
		assert.NoError(t, err)

		patterns := ReadGitignore(dir)
		assert.Equal(t, []string{"*.log", "vendor", "build"}, patterns)
	})

	t.Run("covers close error log path", func(t *testing.T) {
		origOpen := openFile
		openFile = func(name string) (io.ReadCloser, error) {
			return &errCloseRC{r: strings.NewReader("*.log\n")}, nil
		}
		defer func() { openFile = origOpen }()

		patterns := ReadGitignore("/fake/dir")
		assert.Equal(t, []string{"*.log"}, patterns)
	})
}

// fileEntry is a DirEntry that returns IsDir()=false with zero ModTime.
type fileEntry struct{}

func (fileEntry) Name() string               { return "file.go" }
func (fileEntry) IsDir() bool                { return false }
func (fileEntry) Type() fs.FileMode          { return 0 }
func (fileEntry) Info() (fs.FileInfo, error) { return fakeFileInfo{}, nil }

// dirEntry is a DirEntry that returns IsDir()=true with zero ModTime.
type dirEntry struct{}

func (dirEntry) Name() string               { return "dir" }
func (dirEntry) IsDir() bool                { return true }
func (dirEntry) Type() fs.FileMode          { return fs.ModeDir }
func (dirEntry) Info() (fs.FileInfo, error) { return fakeFileInfo{}, nil }

// errEntry is a DirEntry that returns IsDir()=false and Info() returns an error.
type errEntry struct{}

func (errEntry) Name() string               { return "file.go" }
func (errEntry) IsDir() bool                { return false }
func (errEntry) Type() fs.FileMode          { return 0 }
func (errEntry) Info() (fs.FileInfo, error) { return nil, errors.New("stat error") }

func TestWalkFunc(t *testing.T) {
	t.Run("should skip .git directory", func(t *testing.T) {
		root := "/user/project"
		walk := walkFunc(root, time.Now(), nil)

		fi := info{}

		err := walk("/user/project/.git", fi, nil) //nolint:errcheck

		assert.Equal(t, filepath.SkipDir, err, "should Skip directory .git but it not.")
	})

	t.Run("should not skip root even when its name matches an ignore pattern", func(t *testing.T) {
		// Simulates a project directory named "re" with a .gitignore entry "re".
		// The root must never be SkipDir'd or the entire walk is skipped.
		root := "/user/project/re"
		// Use time.Now() so the root's zero ModTime is not "after" lastMod,
		// meaning walkFunc returns nil (no match, no skip) — not SkipDir.
		walk := walkFunc(root, time.Now(), []string{"re"})

		fi := info{} // IsDir() = true

		err := walk(root, fi, nil)

		assert.NoError(t, err, "root directory must not be skipped by an ignore pattern")
	})

	t.Run("non-nil err arg returns nil", func(t *testing.T) {
		root := "/user/project"
		wf := walkFunc(root, time.Now(), nil)
		err := wf("/user/project/somefile", fileEntry{}, errors.New("some fs error"))
		assert.NoError(t, err)
	})

	t.Run("hidden non-directory file returns nil", func(t *testing.T) {
		root := "/user/project"
		wf := walkFunc(root, time.Now(), nil)
		err := wf(".hidden", fileEntry{}, nil)
		assert.NoError(t, err)
	})

	t.Run("pattern-matched non-directory file returns nil", func(t *testing.T) {
		root := "/user/project"
		wf := walkFunc(root, time.Now(), []string{"*.log"})
		err := wf("/user/project/app.log", fileEntry{}, nil)
		assert.NoError(t, err)
	})

	t.Run("d.Info() error returns nil", func(t *testing.T) {
		root := "/user/project"
		wf := walkFunc(root, time.Now(), nil)
		err := wf("/user/project/file.go", errEntry{}, nil)
		assert.NoError(t, err)
	})

	t.Run("hidden directory (non-.git) returns filepath.SkipDir", func(t *testing.T) {
		root := "/user/project"
		wf := walkFunc(root, time.Now(), nil)
		err := wf("/user/project/.hidden_dir", dirEntry{}, nil)
		assert.Equal(t, filepath.SkipDir, err)
	})

	t.Run("pattern-matched directory returns filepath.SkipDir", func(t *testing.T) {
		root := "/user/project"
		wf := walkFunc(root, time.Now(), []string{"vendor"})
		err := wf("/user/project/vendor", dirEntry{}, nil)
		assert.Equal(t, filepath.SkipDir, err)
	})
}
