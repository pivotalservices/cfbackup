package cfbackup

import "github.com/pivotalservices/cfops/tileregistry"

//ElasticRuntimeBuilder -- an object that can build an elastic runtime pre-initialized
type ElasticRuntimeBuilder struct{}

//New -- method to generate an initialized elastic runtime
func (s *ElasticRuntimeBuilder) New(tileSpec tileregistry.TileSpec) (elasticRuntime tileregistry.Tile, err error) {
	elasticRuntime = NewElasticRuntime("", tileSpec.ArchiveDirectory)
	return
}
