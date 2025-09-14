package models

type Fingerprint struct {
	ID         int64  `json:"id"`
	SongID     int64  `json:"song_id"`     // Foreign key to Songs
	Hash       uint32 `json:"hash"`        // Fingerprint hash value // TODO: add index on Hash
	TimeOffset uint32 `json:"time_offset"` // Time offset
}
