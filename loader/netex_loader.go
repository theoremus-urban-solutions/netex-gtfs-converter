package loader

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

// DefaultNetexDatasetLoader implements NetexDatasetLoader
type DefaultNetexDatasetLoader struct{}

// NewDefaultNetexDatasetLoader creates a new default NeTEx dataset loader
func NewDefaultNetexDatasetLoader() producer.NetexDatasetLoader {
	return &DefaultNetexDatasetLoader{}
}

// Load loads NeTEx data from a reader (ZIP archive) into the repository
func (l *DefaultNetexDatasetLoader) Load(data io.Reader, repository producer.NetexRepository) error {
	// Read all data into memory
	zipData, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read ZIP data: %w", err)
	}

	// Open ZIP archive
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
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

		fmt.Printf("Debug: Processing file %s with %d bytes\n", file.Name, len(xmlData))

		// Parse XML and load into repository
		if err := l.parseAndLoadXML(xmlData, repository); err != nil {
			return fmt.Errorf("failed to parse XML file %s: %w", file.Name, err)
		}

		// Also try to parse Network entities from GeneralFrames
		if err := l.parseNetworksFromXML(xmlData, repository); err != nil {
			// Don't fail the whole load if Network parsing fails - log and continue
			fmt.Printf("Warning: Failed to parse Networks from %s: %v\n", file.Name, err)
		}
	}

	return nil
}

// parseAndLoadXML parses NeTEx XML data and loads it into the repository
func (l *DefaultNetexDatasetLoader) parseAndLoadXML(xmlData []byte, repository producer.NetexRepository) error {
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

	// Load data from different frame types
	if err := l.loadResourceFrame(frames.ResourceFrame, repository); err != nil {
		return fmt.Errorf("failed to load resource frame: %w", err)
	}

	if err := l.loadServiceFrame(frames.ServiceFrame, repository); err != nil {
		return fmt.Errorf("failed to load service frame: %w", err)
	}

	if err := l.loadServiceCalendarFrame(frames.ServiceCalendarFrame, repository); err != nil {
		return fmt.Errorf("failed to load service calendar frame: %w", err)
	}

	if err := l.loadTimetableFrame(frames.TimetableFrame, repository); err != nil {
		return fmt.Errorf("failed to load timetable frame: %w", err)
	}

	if err := l.loadSiteFrame(frames.SiteFrame, repository); err != nil {
		return fmt.Errorf("failed to load site frame: %w", err)
	}

	return nil
}

// parseNetworksFromXML extracts Network entities from XML using XML decoder
func (l *DefaultNetexDatasetLoader) parseNetworksFromXML(xmlData []byte, repository producer.NetexRepository) error {
	// Use XML decoder to find Network elements in the raw XML
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	fmt.Printf("Debug: Starting to parse Networks from XML of size %d bytes\n", len(xmlData))

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if startElement, ok := token.(xml.StartElement); ok && startElement.Name.Local == "Network" {
			// Found a Network element - parse it
			var network model.Network
			if err := decoder.DecodeElement(&network, &startElement); err != nil {
				fmt.Printf("Warning: Failed to decode Network element: %v\n", err)
				continue // Skip invalid networks
			}

			memberCount := 0
			if network.Members != nil {
				memberCount = len(network.Members.LineRef)
			}
			fmt.Printf("Debug: Loaded Network ID=%s, AuthorityRef=%s, Members=%d\n",
				network.ID, network.AuthorityRef.Ref, memberCount)

			// Save the network to repository
			if err := repository.SaveEntity(&network); err != nil {
				fmt.Printf("Warning: Failed to save Network %s: %v\n", network.ID, err)
				continue // Skip networks that can't be saved
			}
		}
	}

	return nil
}

// loadResourceFrame loads authorities and other resources
func (l *DefaultNetexDatasetLoader) loadResourceFrame(frame *model.ResourceFrame, repository producer.NetexRepository) error {
	if frame == nil {
		return nil
	}

	// Load authorities
	if frame.Authorities != nil {
		for _, authority := range frame.Authorities.Authority {
			if err := repository.SaveEntity(&authority); err != nil {
				return fmt.Errorf("failed to save authority %s: %w", authority.ID, err)
			}
		}
	}

	return nil
}

// loadServiceFrame loads lines, routes, journey patterns, etc.
func (l *DefaultNetexDatasetLoader) loadServiceFrame(frame *model.ServiceFrame, repository producer.NetexRepository) error {
	if frame == nil {
		return nil
	}

	// Load lines
	if frame.Lines != nil {
		for _, line := range frame.Lines.Line {
			if err := repository.SaveEntity(&line); err != nil {
				return fmt.Errorf("failed to save line %s: %w", line.ID, err)
			}
		}
	}

	// Load routes
	if frame.Routes != nil {
		for _, route := range frame.Routes.Route {
			if err := repository.SaveEntity(&route); err != nil {
				return fmt.Errorf("failed to save route %s: %w", route.ID, err)
			}
		}
	}

	// Load journey patterns
	if frame.JourneyPatterns != nil {
		for _, journeyPattern := range frame.JourneyPatterns.JourneyPattern {
			if err := repository.SaveEntity(&journeyPattern); err != nil {
				return fmt.Errorf("failed to save journey pattern %s: %w", journeyPattern.ID, err)
			}
		}
	}

	// Load destination displays
	if frame.DestinationDisplays != nil {
		for _, destDisplay := range frame.DestinationDisplays.DestinationDisplay {
			if err := repository.SaveEntity(&destDisplay); err != nil {
				return fmt.Errorf("failed to save destination display %s: %w", destDisplay.ID, err)
			}
		}
	}

	// Load scheduled stop points
	if frame.ScheduledStopPoints != nil {
		for _, stopPoint := range frame.ScheduledStopPoints.ScheduledStopPoint {
			if err := repository.SaveEntity(&stopPoint); err != nil {
				return fmt.Errorf("failed to save scheduled stop point %s: %w", stopPoint.ID, err)
			}
		}
	}

	// Load service journey interchanges
	if frame.ServiceJourneyInterchanges != nil {
		for _, interchange := range frame.ServiceJourneyInterchanges.ServiceJourneyInterchange {
			if err := repository.SaveEntity(&interchange); err != nil {
				return fmt.Errorf("failed to save service journey interchange %s: %w", interchange.ID, err)
			}
		}
	}

	return nil
}

// loadServiceCalendarFrame loads day types, operating periods, etc.
func (l *DefaultNetexDatasetLoader) loadServiceCalendarFrame(frame *model.ServiceCalendarFrame, repository producer.NetexRepository) error {
	if frame == nil {
		return nil
	}

	// Load day types
	if frame.DayTypes != nil {
		for _, dayType := range frame.DayTypes.DayType {
			if err := repository.SaveEntity(&dayType); err != nil {
				return fmt.Errorf("failed to save day type %s: %w", dayType.ID, err)
			}
		}
	}

	// Load operating days
	if frame.OperatingDays != nil {
		for _, operatingDay := range frame.OperatingDays.OperatingDay {
			if err := repository.SaveEntity(&operatingDay); err != nil {
				return fmt.Errorf("failed to save operating day %s: %w", operatingDay.ID, err)
			}
		}
	}

	// Load operating periods
	if frame.OperatingPeriods != nil {
		for _, operatingPeriod := range frame.OperatingPeriods.OperatingPeriod {
			if err := repository.SaveEntity(&operatingPeriod); err != nil {
				return fmt.Errorf("failed to save operating period %s: %w", operatingPeriod.ID, err)
			}
		}
	}

	// Load day type assignments
	if frame.DayTypeAssignments != nil {
		for _, dayTypeAssignment := range frame.DayTypeAssignments.DayTypeAssignment {
			if err := repository.SaveEntity(&dayTypeAssignment); err != nil {
				return fmt.Errorf("failed to save day type assignment %s: %w", dayTypeAssignment.ID, err)
			}
		}
	}

	return nil
}

// loadTimetableFrame loads service journeys and timetable data
func (l *DefaultNetexDatasetLoader) loadTimetableFrame(frame *model.TimetableFrame, repository producer.NetexRepository) error {
	if frame == nil {
		return nil
	}

	// Load service journeys
	if frame.ServiceJourneys != nil {
		for _, serviceJourney := range frame.ServiceJourneys.ServiceJourney {
			if err := repository.SaveEntity(&serviceJourney); err != nil {
				return fmt.Errorf("failed to save service journey %s: %w", serviceJourney.ID, err)
			}
		}
	}

	// Load dated service journeys
	if frame.DatedServiceJourneys != nil {
		for _, datedServiceJourney := range frame.DatedServiceJourneys.DatedServiceJourney {
			if err := repository.SaveEntity(&datedServiceJourney); err != nil {
				return fmt.Errorf("failed to save dated service journey %s: %w", datedServiceJourney.ID, err)
			}
		}
	}

	return nil
}

// loadSiteFrame loads stop places and quays
func (l *DefaultNetexDatasetLoader) loadSiteFrame(frame *model.SiteFrame, repository producer.NetexRepository) error {
	if frame == nil {
		return nil
	}

	// Load stop places
	if frame.StopPlaces != nil {
		for _, stopPlace := range frame.StopPlaces.StopPlace {
			if err := repository.SaveEntity(&stopPlace); err != nil {
				return fmt.Errorf("failed to save stop place %s: %w", stopPlace.ID, err)
			}

			// Load quays within stop places
			if stopPlace.Quays != nil {
				for _, quay := range stopPlace.Quays.Quay {
					if err := repository.SaveEntity(&quay); err != nil {
						return fmt.Errorf("failed to save quay %s: %w", quay.ID, err)
					}
				}
			}
		}
	}

	return nil
}
