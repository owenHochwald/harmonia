package models

type Fingerprint struct {
	SongID int64  `json:"song_id"` // Foreign key to Songs
	Hash   uint32 `json:"hash"`    // Fingerprint hash value // TODO: add index on Hash
	Offset uint32 `json:"offset"`  // Time offset
}
