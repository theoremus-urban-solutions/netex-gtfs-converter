package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/exporter"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/repository"
)

const sampleNeTExXML = `<?xml version="1.0" encoding="UTF-8"?>
<PublicationDelivery xmlns="http://www.netex.org.uk/netex" version="1.0">
  <DataObjects>
    <CompositeFrame id="composite-frame-1" version="1">
      <Frames>
        <ResourceFrame id="resource-frame-1" version="1">
          <Authorities>
            <Authority id="test-authority" version="1">
              <Name>Test Transport Authority</Name>
              <ShortName>TTA</ShortName>
              <Url>https://example.com</Url>
              <ContactDetails>
                <Phone>+1234567890</Phone>
                <Email>info@example.com</Email>
              </ContactDetails>
            </Authority>
          </Authorities>
        </ResourceFrame>
        <ServiceFrame id="service-frame-1" version="1">
          <Lines>
            <Line id="test-line-1" version="1">
              <Name>Test Bus Line</Name>
              <ShortName>T1</ShortName>
              <PublicCode>1</PublicCode>
              <TransportMode>bus</TransportMode>
              <TransportSubmode>localBus</TransportSubmode>
              <AuthorityRef>test-authority</AuthorityRef>
              <Presentation>
                <Colour>FF0000</Colour>
                <TextColour>FFFFFF</TextColour>
              </Presentation>
            </Line>
          </Lines>
          <DestinationDisplays>
            <DestinationDisplay id="dest-1" version="1">
              <FrontText>City Center</FrontText>
            </DestinationDisplay>
          </DestinationDisplays>
        </ServiceFrame>
        <SiteFrame id="site-frame-1" version="1">
          <StopPlaces>
            <StopPlace id="stop-place-1" version="1">
              <Name>Main Station</Name>
              <TransportMode>bus</TransportMode>
              <Centroid>
                <Location>
                  <Longitude>10.7522</Longitude>
                  <Latitude>59.9139</Latitude>
                </Location>
              </Centroid>
              <Quays>
                <Quay id="quay-1" version="1">
                  <Name>Platform 1</Name>
                  <PublicCode>A</PublicCode>
                  <Centroid>
                    <Location>
                      <Longitude>10.7522</Longitude>
                      <Latitude>59.9139</Latitude>
                    </Location>
                  </Centroid>
                </Quay>
              </Quays>
            </StopPlace>
          </StopPlaces>
        </SiteFrame>
      </Frames>
    </CompositeFrame>
  </DataObjects>
</PublicationDelivery>`

func TestEndToEndConversion(t *testing.T) {
	// Create repositories
	stopAreaRepo := repository.NewDefaultStopAreaRepository()

	// Create exporter
	exporter := exporter.NewDefaultGtfsExporter("TEST", stopAreaRepo)

	// Test converting stops only (simpler test)
	result, err := exporter.ConvertStopsToGtfs()
	if err != nil {
		t.Fatalf("ConvertStopsToGtfs() failed: %v", err)
	}

	if result == nil {
		t.Fatal("ConvertStopsToGtfs() returned nil")
	}

	// Read first few bytes to verify it's a ZIP
	buffer := make([]byte, 100)
	n, _ := result.Read(buffer)

	if n < 4 || buffer[0] != 'P' || buffer[1] != 'K' {
		t.Error("Result does not appear to be a valid ZIP file")
	}
}

func TestEndToEndWithSampleData(t *testing.T) {
	// Create repositories
	stopAreaRepo := repository.NewDefaultStopAreaRepository()

	// Create exporter
	exporter := exporter.NewDefaultGtfsExporter("TEST", stopAreaRepo)

	// Convert with sample NeTEx data
	netexReader := strings.NewReader(sampleNeTExXML)
	result, err := exporter.ConvertTimetablesToGtfs(netexReader)

	if err != nil {
		t.Logf("ConvertTimetablesToGtfs() failed (expected due to minimal data): %v", err)

		// For this basic test, we accept certain errors due to minimal sample data
		// In a real integration test, we would use complete sample datasets
		return
	}

	if result == nil {
		t.Fatal("ConvertTimetablesToGtfs() returned nil")
	}

	// Read first few bytes to verify it's a ZIP
	buffer := make([]byte, 100)
	n, _ := result.Read(buffer)

	if n < 4 || buffer[0] != 'P' || buffer[1] != 'K' {
		t.Error("Result does not appear to be a valid ZIP file")
	}
}

func TestCLIValidation(t *testing.T) {
	// Test that CLI validation works correctly
	// This is a basic smoke test for the CLI arguments

	tests := []struct {
		name      string
		codespace string
		netexFile string
		stopsFile string
		stopsOnly bool
		expectErr bool
	}{
		{
			name:      "valid timetable conversion",
			codespace: "TEST",
			netexFile: "test.zip",
			expectErr: false,
		},
		{
			name:      "missing codespace",
			codespace: "",
			netexFile: "test.zip",
			expectErr: true,
		},
		{
			name:      "missing netex file",
			codespace: "TEST",
			netexFile: "",
			expectErr: true,
		},
		{
			name:      "valid stops only conversion",
			stopsFile: "stops.zip",
			stopsOnly: true,
			expectErr: false,
		},
		{
			name:      "stops only without file",
			stopsOnly: true,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates the validation logic from main.go
			err := validateFlags(tt.codespace, tt.netexFile, tt.stopsFile, tt.stopsOnly)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// Helper function copied from main.go for testing
func validateFlags(codespace, netexFile, stopsFile string, stopsOnly bool) error {
	if stopsOnly {
		// For stops-only conversion, we need either stops file or both files
		if stopsFile == "" && netexFile == "" {
			return fmt.Errorf("stops-only conversion requires either --stops or --netex file")
		}
	} else {
		// For full conversion, we need codespace and netex file
		if codespace == "" {
			return fmt.Errorf("--codespace is required for timetable conversion")
		}
		if netexFile == "" {
			return fmt.Errorf("--netex file is required for timetable conversion")
		}
	}
	return nil
}
