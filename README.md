# Harmonia 

A Shazam-like music identification service built with Go, implementing audio fingerprinting and spectral analysis for real-time song recognition.

## Architecture

**Backend Stack:**
- **API Framework:** Go + Gin (RESTful API)
- **Architecture:** MVC + Clean Architecture patterns
- **Database:** PostgreSQL (song metadata + audio fingerprints)
- **Storage:** AWS S3 (raw audio files)
- **Audio Processing:** FFT/STFT spectral analysis
- **Deployment:** AWS ECS + Docker

## How It Works

1. **Audio Upload** → Raw audio stored in S3
2. **Spectral Analysis** → FFT generates frequency domain representation
3. **Peak Detection** → Identifies prominent frequency peaks over time
4. **Fingerprint Generation** → Creates hash signatures from peak constellation pairs
5. **Database Storage** → Stores hashes with time offsets for fast retrieval
6. **Song Identification** → Queries fingerprint database for matches

## API Endpoints

```
POST /upload     - Upload and fingerprint audio files
POST /identify   - Identify song from audio sample
GET  /health     - Service health check
```

## Project Structure

```
cmd/api/           # Application entry point
internal/
├── config/        # Configuration management
├── server/        # HTTP handlers and routing
├── models/        # Data models (Song, Fingerprint)
├── services/      # Business logic layer
├── repo/          # Database repository interfaces
└── storage/       # S3 storage interface
pkg/logger/        # Structured logging
```

## Development Setup

**Prerequisites:**
- Go 1.21+
- Docker & Docker Compose
- AWS CLI (for S3 integration)

**Quick Start:**
```bash
# Clone and setup
git clone https://github.com/owenHochwald/harmonia
cd harmonia

# Start local database
docker-compose up -d

# Install dependencies
go mod tidy

# Run application
go run cmd/api/main.go
```

## Technical Implementation

**Audio Fingerprinting Algorithm:**
- **Spectrogram Generation:** STFT with overlapping windows
- **Peak Detection:** Local maxima identification in time-frequency domain
- **Constellation Mapping:** Anchor-target peak pair generation
- **Hash Function:** `hash(freq1, freq2, time_delta)` for unique signatures

**Database Schema:**
```sql
songs (id, title, artist, album, s3_key, created_at)
fingerprints (song_id, hash, offset) -- Indexed on hash for O(log n) lookup
```

**Performance Optimizations:**
- Database indexing on fingerprint hashes
- Efficient hash collision handling
- Time-offset consistency validation for match scoring

## Deployment

**AWS Infrastructure:**
- **ECS Fargate:** Containerized API deployment
- **RDS PostgreSQL:** Managed database with read replicas
- **S3:** Audio file storage with lifecycle policies
- **CloudWatch:** Logging and monitoring

## Testing

**Separate Test Database:**
- Development DB: `localhost:5432` (docker-compose.yml)
- Test DB: `localhost:5433` (docker-compose.test.yml)

**Quick Start:**
```bash
# Setup and run all tests
./scripts/test-setup.sh run

# Manual setup
docker-compose -f docker-compose.test.yml up -d
go test ./internal/repo -v
```

**Test Types:**
```bash
# Unit tests with mocks (no database required)
go test ./internal/repo -run TestMock -v

# Integration tests (requires test database)
docker-compose -f docker-compose.test.yml up -d
go test ./internal/repo -v

# Specific test patterns
go test ./internal/repo -run TestSongRepo_SaveSong -v
go test ./internal/repo -run TestFingerprintRepo -v
```

**Test Database Management:**
```bash
# Start test database
docker-compose -f docker-compose.test.yml up -d

# Stop test database
docker-compose -f docker-compose.test.yml down

# Reset test database (remove volumes)
docker-compose -f docker-compose.test.yml down -v
```

## Future Enhancements

- **OpenSearch Integration:** Scale fingerprint search for large datasets
- **Real-time Processing:** WebSocket-based live audio identification
- **Machine Learning:** Neural network-based audio feature extraction
- **Multi-format Support:** AAC, FLAC, OGG audio format compatibility

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## Learning Resources (so you can make it yourself!)
- [Coding Geek: How does Shazaam work?](https://drive.google.com/file/d/1ahyCTXBAZiuni6RTzHzLoOwwfTRFaU-C/view)
- [Audio Hashing Paper](https://hajim.rochester.edu/ece/sites/zduan/teaching/ece472/projects/2019/AudioFingerprinting.pdf)
## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.