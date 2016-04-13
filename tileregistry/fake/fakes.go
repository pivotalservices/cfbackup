package fake

import (
	"github.com/pivotalservices/cfbackup/tileregistry"
	"github.com/xchapter7x/lo"
)

//TileGenerator --
type TileGenerator struct {
	tileregistry.TileGenerator
	TileSpy tileregistry.Tile
	ErrFake error
	Closer  tileregistry.Closer
}

//Closer --
type Closer struct {
	Executions int
}

//Close --
func (closer *Closer) Close() {
	closer.Executions++
}

//New --
func (s *TileGenerator) New(tileSpec tileregistry.TileSpec) (tileregistry.TileCloser, error) {
	tileCloser := struct {
		tileregistry.Tile
		tileregistry.Closer
	}{
		s.TileSpy,
		s.Closer,
	}
	return tileCloser, s.ErrFake
}

//Tile --
type Tile struct {
	ErrFake          error
	BackupCallCount  int
	RestoreCallCount int
}

//Backup --
func (s *Tile) Backup() error {
	lo.G.Debug("we fake backed up")
	s.BackupCallCount++
	return s.ErrFake
}

//Restore --
func (s *Tile) Restore() error {
	lo.G.Debug("we fake restored")
	s.RestoreCallCount++
	return s.ErrFake
}
