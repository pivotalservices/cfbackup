package tileregistry

type (
	//TileGenerator - interface for a tile creating object
	TileGenerator interface {
		New(tileSpec TileSpec) (TileCloser, error)
	}
	//Tile - definition for what a tile looks like
	Tile interface {
		Backup() error
		Restore() error
	}

	//Closer - define how to close the tile
	Closer interface {
		Close()
	}

	//TileCloser - defines how to close a tile
	TileCloser interface {
		Tile
		Closer
	}

	//DoNothingCloser - This Closer do nothing
	DoNothingCloser struct {
	}
	//TileSpec -- defines what a tile would need to be initialized
	TileSpec struct {
		OpsManagerHost       string
		AdminUser            string
		AdminPass            string
		OpsManagerUser       string
		OpsManagerPass       string
		OpsManagerPassphrase string
		ArchiveDirectory     string
		CryptKey             string
		ClearBoshManifest    bool
		PluginArgs           string
	}
)
