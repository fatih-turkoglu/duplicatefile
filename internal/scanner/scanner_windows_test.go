package scanner

import (
	"syscall"
	"testing"
)

func TestScanSkipHidden(t *testing.T) {
	root, p := fixture(t)
	hidden := write(t, root, `gizli\dup.txt`, []byte("hello world"))
	name, _ := syscall.UTF16PtrFromString(p["dup3"])
	if err := syscall.SetFileAttributes(name, syscall.FILE_ATTRIBUTE_HIDDEN); err != nil {
		t.Fatal(err)
	}
	dir, _ := syscall.UTF16PtrFromString(root + `\gizli`)
	if err := syscall.SetFileAttributes(dir, syscall.FILE_ATTRIBUTE_HIDDEN|syscall.FILE_ATTRIBUTE_DIRECTORY); err != nil {
		t.Fatal(err)
	}
	s := New(openDB(t))

	res := scan(t, s, Options{Roots: []string{root}, ExcludeDirs: []string{p["haric"]}, SkipHidden: true})
	checkGroups(t, res, [][]string{
		{p["big1"], p["big2"]},
		{p["dup1"], p["dup2"]},
	})

	res = scan(t, s, Options{Roots: []string{root}, ExcludeDirs: []string{p["haric"]}})
	checkGroups(t, res, [][]string{
		{p["big1"], p["big2"]},
		{p["dup1"], p["dup3"], p["dup2"], hidden},
	})
}
