package repository

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// DefaultStopAreaRepository implements StopAreaRepository
type DefaultStopAreaRepository struct {
	stopPlaces        map[string]*model.StopPlace
	quays             map[string]*model.Quay
	stopPlaceByQuayId map[string]*model.StopPlace
}

// NewDefaultStopAreaRepository creates a new DefaultStopAreaRepository
func NewDefaultStopAreaRepository() producer.StopAreaRepository {
	return &DefaultStopAreaRepository{
		stopPlaces:        make(map[string]*model.StopPlace),
		quays:             make(map[string]*model.Quay),
		stopPlaceByQuayId: make(map[string]*model.StopPlace),
	}
}

// GetQuayById returns a quay by ID
func (r *DefaultStopAreaRepository) GetQuayById(quayId string) *model.Quay {
	return r.quays[quayId]
}

// GetStopPlaceByQuayId returns the stop place for a given quay ID
func (r *DefaultStopAreaRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace {
	return r.stopPlaceByQuayId[quayId]
}

// GetAllQuays returns all quays
func (r *DefaultStopAreaRepository) GetAllQuays() []*model.Quay {
	quays := make([]*model.Quay, 0, len(r.quays))
	for _, quay := range r.quays {
		quays = append(quays, quay)
	}
	return quays
}

// LoadStopAreas loads stop area data from a ZIP archive
func (r *DefaultStopAreaRepository) LoadStopAreas(data []byte) error {
	// Open ZIP archive
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("failed to open ZIP archive: %w", err)
	}

	// Process each XML file in the archive
	for _, file := range zipReader.File {
		if !strings.HasSuffix(strings.ToLower(file.Name), ".xml") {
			continue
		}

		// Open file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}

		// Read file content
		xmlData, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		// Parse XML and load stop areas
		if err := r.parseStopAreaXML(xmlData); err != nil {
			return fmt.Errorf("failed to parse stop area XML file %s: %w", file.Name, err)
		}
	}

	return nil
}

// parseStopAreaXML parses stop area XML data and loads it into the repository
func (r *DefaultStopAreaRepository) parseStopAreaXML(xmlData []byte) error {
	// Parse the root PublicationDelivery structure
	var pubDelivery model.PublicationDelivery
	if err := xml.Unmarshal(xmlData, &pubDelivery); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// Extract frames from the structure
	var compositeFrame *model.CompositeFrame
	if pubDelivery.DataObjects != nil && pubDelivery.DataObjects.CompositeFrame != nil {
		compositeFrame = pubDelivery.DataObjects.CompositeFrame
	} else if pubDelivery.CompositeFrame != nil {
		compositeFrame = pubDelivery.CompositeFrame
	}

	if compositeFrame == nil {
		return fmt.Errorf("no CompositeFrame found in XML")
	}

	if compositeFrame.Frames == nil {
		return fmt.Errorf("no Frames found in CompositeFrame")
	}

	frames := compositeFrame.Frames

	// Load stop areas from SiteFrame
	if frames.SiteFrame != nil {
		if err := r.loadStopAreasFromSiteFrame(frames.SiteFrame); err != nil {
			return fmt.Errorf("failed to load stop areas from site frame: %w", err)
		}
	}

	return nil
}

// loadStopAreasFromSiteFrame loads stop places and quays from a SiteFrame
func (r *DefaultStopAreaRepository) loadStopAreasFromSiteFrame(siteFrame *model.SiteFrame) error {
	if siteFrame.StopPlaces == nil {
		return nil
	}

	// Load stop places and their quays
	for _, stopPlace := range siteFrame.StopPlaces.StopPlace {
		// Store the stop place
		r.stopPlaces[stopPlace.ID] = &stopPlace

		// Load quays within the stop place
		if stopPlace.Quays != nil {
			for _, quay := range stopPlace.Quays.Quay {
				// Store the quay
				r.quays[quay.ID] = &quay

				// Map quay to its stop place
				r.stopPlaceByQuayId[quay.ID] = &stopPlace
			}
		}
	}

	return nil
}

// GetStopPlaceById returns a stop place by ID (helper method)
func (r *DefaultStopAreaRepository) GetStopPlaceById(id string) *model.StopPlace {
	return r.stopPlaces[id]
}

// GetAllStopPlaces returns all stop places (helper method)
func (r *DefaultStopAreaRepository) GetAllStopPlaces() []*model.StopPlace {
	stopPlaces := make([]*model.StopPlace, 0, len(r.stopPlaces))
	for _, stopPlace := range r.stopPlaces {
		stopPlaces = append(stopPlaces, stopPlace)
	}
	return stopPlaces
}
