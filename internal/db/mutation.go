package db

// CatalogPlan is a point-in-time review of one global catalog operation.
// Fingerprint is opaque and must be checked again in the write transaction.
type CatalogPlan struct {
	Action, Kind                                                                               string
	From, To                                                                                   int32
	ReplaceImage                                                                               bool
	SourceName, DestinationName, SourceImage, DestinationImage                                 string
	SourceTracks, SourceAlbums, SourceListens, RemovedTracks, RemovedListens, DuplicateListens int64
	DestinationProfiles                                                                        []CatalogImpactProfile
	Profiles                                                                                   []CatalogImpactProfile
	SourceMBID, DestinationMBID                                                                string
	SourceAliases, DestinationAliases                                                          []string
	SourceArtists, DestinationArtists                                                          []CatalogImpactArtist
	RemovedAlbums                                                                              int64
	AffectedTrackIDs                                                                           []int32
	Fingerprint                                                                                string
}
type CatalogImpactProfile struct {
	ID      int32
	Name    string
	Listens int64
}

type CatalogImpactArtist struct {
	ID      int32
	Name    string
	Primary bool
}
