package elasticruntime

import (
	"io/ioutil"
	"os"

	"github.com/xchapter7x/lo"
)

//NewTempFile - creates a temporary file which can be cleaned up
func NewTempFile(filename string) (*TempFile, error) {
	installationTmpFile, err := ioutil.TempFile("", filename)
	return &TempFile{
		FileRef: installationTmpFile,
	}, err
}

//Close - allows us to clean up the tmp file (delete, close)
func (s *TempFile) Close() {
	lo.G.Debug("removing tmpfile: ", s.FileRef.Name())
	s.FileRef.Close()

	if e := os.Remove(s.FileRef.Name()); e != nil {
		lo.G.Errorf("removing tmpfile: %s failed with %s", s.FileRef.Name(), e)
	}
}
