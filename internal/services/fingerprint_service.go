package services

import (
	"github.com/owenhochwald/harmonia/internal/models"
	"github.com/owenhochwald/harmonia/internal/repo"
)

type FingerprintServiceInterface interface {
	GenerateFingerprints(spec *Spectrogram) ([]models.Fingerprint, error)
}

type FingerprintService struct {
	Repo            repo.FingerprintRepo
	lowBandMax      int
	midBandMax      int
	targetZone      int
	maxPairsPerPeak int
	peakThreshold   float64
}

type Peak struct {
	TimeFrame int
	FreqBin   int
	Magnitude float64
}

type LandmarkPair struct {
	Freq1      int
	Freq2      int
	TimeDelta  int
	AnchorTime int
}

func NewFingerprintService(repo repo.FingerprintRepo) FingerprintServiceInterface {
	return &FingerprintService{
		Repo:            repo,
		lowBandMax:      64,
		midBandMax:      256,
		targetZone:      5,
		maxPairsPerPeak: 5,
		peakThreshold:   1.5,
	}
}

func (f *FingerprintService) FindPeaks(spec *Spectrogram) []Peak {
	var peaks []Peak

	numFrames := len(spec.Data)
	if numFrames == 0 {
		return peaks
	}

	numBins := len(spec.Data[0])
	highBandMax := numBins

	for frameIdx := 0; frameIdx < numFrames; frameIdx++ {
		frame := spec.Data[frameIdx]

		sum := 0.0
		for _, mag := range frame {
			sum += mag
		}
		mean := sum / float64(len(frame))
		threshold := mean * f.peakThreshold

		bands := []struct{ start, end int }{
			{0, f.lowBandMax},
			{f.lowBandMax, f.midBandMax},
			{f.midBandMax, highBandMax},
		}

		for _, band := range bands {
			maxMag := 0.0
			maxBin := -1

			for bin := band.start; bin < band.end && bin < len(frame); bin++ {
				if frame[bin] > maxMag && frame[bin] > threshold {
					maxMag = frame[bin]
					maxBin = bin
				}
			}

			if maxBin != -1 {
				peaks = append(peaks, Peak{
					TimeFrame: frameIdx,
					FreqBin:   maxBin,
					Magnitude: maxMag,
				})
			}
		}
	}

	return peaks
}

func (f *FingerprintService) CreateLandmarkPairs(peaks []Peak) []LandmarkPair {
	var pairs []LandmarkPair

	peaksByFrame := make(map[int][]Peak)
	for _, peak := range peaks {
		peaksByFrame[peak.TimeFrame] = append(peaksByFrame[peak.TimeFrame], peak)
	}

	for _, anchor := range peaks {
		pairCount := 0

		for t := anchor.TimeFrame + 1; t <= anchor.TimeFrame+f.targetZone; t++ {
			targetPeaks, exists := peaksByFrame[t]
			if !exists {
				continue
			}

			for _, target := range targetPeaks {
				if pairCount >= f.maxPairsPerPeak {
					break
				}

				pairs = append(pairs, LandmarkPair{
					Freq1:      anchor.FreqBin,
					Freq2:      target.FreqBin,
					TimeDelta:  target.TimeFrame - anchor.TimeFrame,
					AnchorTime: anchor.TimeFrame,
				})
				pairCount++
			}

			if pairCount >= f.maxPairsPerPeak {
				break
			}
		}
	}

	return pairs
}

func (f *FingerprintService) HashPair(pair LandmarkPair) uint32 {
	freq1 := uint32(pair.Freq1)
	if freq1 > 4095 {
		freq1 = 4095
	}

	freq2 := uint32(pair.Freq2)
	if freq2 > 1023 {
		freq2 = 1023
	}

	timeDelta := uint32(pair.TimeDelta)
	if timeDelta > 1023 {
		timeDelta = 1023
	}

	hash := (freq1 << 20) | (freq2 << 10) | timeDelta
	return hash
}

func (f *FingerprintService) GenerateFingerprints(spec *Spectrogram) ([]models.Fingerprint, error) {
	peaks := f.FindPeaks(spec)

	if len(peaks) == 0 {
		return []models.Fingerprint{}, nil // Silent audio or no peaks found
	}

	pairs := f.CreateLandmarkPairs(peaks)

	if len(pairs) == 0 {
		return []models.Fingerprint{}, nil // No pairs created
	}

	fingerprints := make([]models.Fingerprint, 0, len(pairs))

	for _, pair := range pairs {
		hash := f.HashPair(pair)

		fingerprints = append(fingerprints, models.Fingerprint{
			Hash:       hash,
			TimeOffset: uint32(pair.AnchorTime),
			// SongID will be set by MusicService
		})
	}

	return fingerprints, nil
}
