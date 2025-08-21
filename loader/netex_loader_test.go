package loader

import (
	"bytes"
	"strings"
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

// mockNetexRepository implements NetexRepository for testing
type mockNetexRepository struct {
	entities []interface{}
	saveErr  error
}

func (m *mockNetexRepository) SaveEntity(entity interface{}) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.entities = append(m.entities, entity)
	return nil
}

func (m *mockNetexRepository) GetLines() []*model.Line {
	lines := make([]*model.Line, 0)
	for _, entity := range m.entities {
		if line, ok := entity.(*model.Line); ok {
			lines = append(lines, line)
		}
	}
	return lines
}

func (m *mockNetexRepository) GetAuthorities() []*model.Authority {
	authorities := make([]*model.Authority, 0)
	for _, entity := range m.entities {
		if authority, ok := entity.(*model.Authority); ok {
			authorities = append(authorities, authority)
		}
	}
	return authorities
}

func (m *mockNetexRepository) GetRoutes() []*model.Route {
	routes := make([]*model.Route, 0)
	for _, entity := range m.entities {
		if route, ok := entity.(*model.Route); ok {
			routes = append(routes, route)
		}
	}
	return routes
}

func (m *mockNetexRepository) GetServiceJourneys() []*model.ServiceJourney {
	journeys := make([]*model.ServiceJourney, 0)
	for _, entity := range m.entities {
		if journey, ok := entity.(*model.ServiceJourney); ok {
			journeys = append(journeys, journey)
		}
	}
	return journeys
}

func (m *mockNetexRepository) GetStopPlaces() []*model.StopPlace {
	stops := make([]*model.StopPlace, 0)
	for _, entity := range m.entities {
		if stop, ok := entity.(*model.StopPlace); ok {
			stops = append(stops, stop)
		}
	}
	return stops
}

func (m *mockNetexRepository) GetQuays() []*model.Quay {
	quays := make([]*model.Quay, 0)
	for _, entity := range m.entities {
		if quay, ok := entity.(*model.Quay); ok {
			quays = append(quays, quay)
		}
	}
	return quays
}

func (m *mockNetexRepository) GetQuayById(id string) *model.Quay {
	for _, entity := range m.entities {
		if quay, ok := entity.(*model.Quay); ok && quay.ID == id {
			return quay
		}
	}
	return nil
}

func (m *mockNetexRepository) GetStopPlaceByQuayId(quayId string) *model.StopPlace {
	for _, entity := range m.entities {
		if stopPlace, ok := entity.(*model.StopPlace); ok && stopPlace.ID == quayId {
			return stopPlace
		}
	}
	return nil
}

func (m *mockNetexRepository) GetJourneyPatterns() map[string]*model.JourneyPattern {
	patterns := make(map[string]*model.JourneyPattern)
	for _, entity := range m.entities {
		if pattern, ok := entity.(*model.JourneyPattern); ok {
			patterns[pattern.ID] = pattern
		}
	}
	return patterns
}

func (m *mockNetexRepository) GetAuthorityById(id string) *model.Authority {
	for _, entity := range m.entities {
		if authority, ok := entity.(*model.Authority); ok && authority.ID == id {
			return authority
		}
	}
	return nil
}

func (m *mockNetexRepository) GetAuthorityIdForLine(line *model.Line) string {
	return line.AuthorityRef
}

func (m *mockNetexRepository) GetTimeZone() string {
	return "Europe/Oslo"
}

func (m *mockNetexRepository) GetAllQuays() []*model.Quay {
	return m.GetQuays()
}

func (m *mockNetexRepository) GetJourneyPatternById(id string) *model.JourneyPattern {
	return m.GetJourneyPatterns()[id]
}

func (m *mockNetexRepository) GetRouteById(id string) *model.Route {
	for _, entity := range m.entities {
		if route, ok := entity.(*model.Route); ok && route.ID == id {
			return route
		}
	}
	return nil
}

func (m *mockNetexRepository) GetRoutesByLine(line *model.Line) []*model.Route {
	var routes []*model.Route
	for _, entity := range m.entities {
		if route, ok := entity.(*model.Route); ok {
			if route.LineRef.Ref == line.ID {
				routes = append(routes, route)
			}
		}
	}
	return routes
}

func (m *mockNetexRepository) GetServiceJourneysByJourneyPattern(pattern *model.JourneyPattern) []*model.ServiceJourney {
	var journeys []*model.ServiceJourney
	for _, entity := range m.entities {
		if journey, ok := entity.(*model.ServiceJourney); ok {
			if journey.JourneyPatternRef.Ref == pattern.ID {
				journeys = append(journeys, journey)
			}
		}
	}
	return journeys
}

func (m *mockNetexRepository) GetServiceJourneyInterchanges() []*model.ServiceJourneyInterchange {
	var interchanges []*model.ServiceJourneyInterchange
	for _, entity := range m.entities {
		if interchange, ok := entity.(*model.ServiceJourneyInterchange); ok {
			interchanges = append(interchanges, interchange)
		}
	}
	return interchanges
}

func (m *mockNetexRepository) GetDestinationDisplayById(id string) *model.DestinationDisplay {
	for _, entity := range m.entities {
		if display, ok := entity.(*model.DestinationDisplay); ok && display.ID == id {
			return display
		}
	}
	return nil
}

func (m *mockNetexRepository) GetDatedServiceJourneysByServiceJourneyId(serviceJourneyId string) []*model.DatedServiceJourney {
	return nil
}

func (m *mockNetexRepository) GetDayTypeById(id string) *model.DayType {
	return nil
}

func (m *mockNetexRepository) GetOperatingDayById(id string) *model.OperatingDay {
	return nil
}

func (m *mockNetexRepository) GetDayTypeAssignmentsByDayType(dayType *model.DayType) []*model.DayTypeAssignment {
	return nil
}

func (m *mockNetexRepository) GetScheduledStopPointRefByPointInJourneyPatternRef(pjpRef string) string {
	return pjpRef
}

func (m *mockNetexRepository) GetStopPointInJourneyPatternById(id string) *model.StopPointInJourneyPattern {
	return nil
}

func (m *mockNetexRepository) GetScheduledStopPointById(id string) *model.ScheduledStopPoint {
	return nil
}

func (m *mockNetexRepository) GetAllStopPlaces() []*model.StopPlace {
	return m.GetStopPlaces()
}

func (m *mockNetexRepository) GetHeadwayJourneyGroups() []*model.HeadwayJourneyGroup {
	return nil
}

func (m *mockNetexRepository) GetHeadwayJourneyGroupById(id string) *model.HeadwayJourneyGroup {
	return nil
}

func TestNewDefaultNetexDatasetLoader(t *testing.T) {
	loader := NewDefaultNetexDatasetLoader()
	if loader == nil {
		t.Fatal("NewDefaultNetexDatasetLoader() returned nil")
	}

	// Check that it implements the interface
	if loader == nil {
		t.Error("DefaultNetexDatasetLoader returned nil")
	}
}

func TestDefaultNetexDatasetLoader_LoadInvalidData(t *testing.T) {
	loader := NewDefaultNetexDatasetLoader()
	repo := &mockNetexRepository{}

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "empty input",
			input:     "",
			expectErr: true,
		},
		{
			name:      "invalid zip data",
			input:     "not a zip file",
			expectErr: true,
		},
		{
			name:      "corrupted zip",
			input:     "PK\x03\x04corrupted",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			err := loader.Load(reader, repo)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultNetexDatasetLoader_ParseAndLoadXML(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	// Test XML with minimal structure
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<PublicationDelivery xmlns="http://www.netex.org.uk/netex">
	<CompositeFrame>
		<Frames>
			<ResourceFrame>
				<Authorities>
					<Authority id="test-authority" version="1">
						<Name>Test Authority</Name>
					</Authority>
				</Authorities>
			</ResourceFrame>
		</Frames>
	</CompositeFrame>
</PublicationDelivery>`

	err := loader.parseAndLoadXML([]byte(xmlData), repo)
	if err != nil {
		t.Fatalf("parseAndLoadXML() failed: %v", err)
	}

	authorities := repo.GetAuthorities()
	if len(authorities) != 1 {
		t.Errorf("Expected 1 authority, got %d", len(authorities))
	}

	if len(authorities) > 0 && authorities[0].ID != "test-authority" {
		t.Errorf("Expected authority ID 'test-authority', got '%s'", authorities[0].ID)
	}
}

func TestDefaultNetexDatasetLoader_ParseAndLoadXMLInvalidStructure(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	tests := []struct {
		name      string
		xmlData   string
		expectErr bool
	}{
		{
			name:      "invalid XML",
			xmlData:   "not valid xml",
			expectErr: true,
		},
		{
			name: "no CompositeFrame",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<PublicationDelivery xmlns="http://www.netex.org.uk/netex">
</PublicationDelivery>`,
			expectErr: true,
		},
		{
			name: "no frames in CompositeFrame",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<PublicationDelivery xmlns="http://www.netex.org.uk/netex">
	<CompositeFrame>
	</CompositeFrame>
</PublicationDelivery>`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.parseAndLoadXML([]byte(tt.xmlData), repo)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultNetexDatasetLoader_LoadResourceFrame(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	// Test with nil frame
	err := loader.loadResourceFrame(nil, repo)
	if err != nil {
		t.Errorf("loadResourceFrame(nil) should not error, got: %v", err)
	}

	// Test with frame containing authorities
	frame := &model.ResourceFrame{
		Authorities: &model.Authorities{
			Authority: []model.Authority{
				{ID: "auth1", Name: "Authority 1"},
				{ID: "auth2", Name: "Authority 2"},
			},
		},
	}

	err = loader.loadResourceFrame(frame, repo)
	if err != nil {
		t.Fatalf("loadResourceFrame() failed: %v", err)
	}

	authorities := repo.GetAuthorities()
	if len(authorities) != 2 {
		t.Errorf("Expected 2 authorities, got %d", len(authorities))
	}
}

func TestDefaultNetexDatasetLoader_LoadServiceFrame(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	// Test with nil frame
	err := loader.loadServiceFrame(nil, repo)
	if err != nil {
		t.Errorf("loadServiceFrame(nil) should not error, got: %v", err)
	}

	// Test with frame containing various entities
	frame := &model.ServiceFrame{
		Lines: &model.Lines{
			Line: []model.Line{
				{ID: "line1", Name: "Line 1"},
			},
		},
		Routes: &model.Routes{
			Route: []model.Route{
				{ID: "route1", Name: "Route 1"},
			},
		},
		JourneyPatterns: &model.JourneyPatterns{
			JourneyPattern: []model.JourneyPattern{
				{ID: "jp1", Name: "Journey Pattern 1"},
			},
		},
	}

	err = loader.loadServiceFrame(frame, repo)
	if err != nil {
		t.Fatalf("loadServiceFrame() failed: %v", err)
	}

	lines := repo.GetLines()
	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	routes := repo.GetRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	patterns := repo.GetJourneyPatterns()
	if len(patterns) != 1 {
		t.Errorf("Expected 1 journey pattern, got %d", len(patterns))
	}
}

func TestDefaultNetexDatasetLoader_LoadTimetableFrame(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	// Test with nil frame
	err := loader.loadTimetableFrame(nil, repo)
	if err != nil {
		t.Errorf("loadTimetableFrame(nil) should not error, got: %v", err)
	}

	// Test with frame containing service journeys
	frame := &model.TimetableFrame{
		ServiceJourneys: &model.ServiceJourneys{
			ServiceJourney: []model.ServiceJourney{
				{ID: "sj1", JourneyPatternRef: model.ServiceJourneyPatternRef{Ref: "jp1"}},
			},
		},
	}

	err = loader.loadTimetableFrame(frame, repo)
	if err != nil {
		t.Fatalf("loadTimetableFrame() failed: %v", err)
	}

	journeys := repo.GetServiceJourneys()
	if len(journeys) != 1 {
		t.Errorf("Expected 1 service journey, got %d", len(journeys))
	}
}

func TestDefaultNetexDatasetLoader_LoadSiteFrame(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	// Test with nil frame
	err := loader.loadSiteFrame(nil, repo)
	if err != nil {
		t.Errorf("loadSiteFrame(nil) should not error, got: %v", err)
	}

	// Test with frame containing stop places and quays
	frame := &model.SiteFrame{
		StopPlaces: &model.StopPlaces{
			StopPlace: []model.StopPlace{
				{
					ID:   "sp1",
					Name: "Stop Place 1",
					Quays: &model.Quays{
						Quay: []model.Quay{
							{ID: "quay1", Name: "Platform 1"},
							{ID: "quay2", Name: "Platform 2"},
						},
					},
				},
			},
		},
	}

	err = loader.loadSiteFrame(frame, repo)
	if err != nil {
		t.Fatalf("loadSiteFrame() failed: %v", err)
	}

	stopPlaces := repo.GetStopPlaces()
	if len(stopPlaces) != 1 {
		t.Errorf("Expected 1 stop place, got %d", len(stopPlaces))
	}

	quays := repo.GetQuays()
	if len(quays) != 2 {
		t.Errorf("Expected 2 quays, got %d", len(quays))
	}
}

func TestDefaultNetexDatasetLoader_ParseNetworksFromXML(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<Network id="net1" version="1">
		<Name>Test Network</Name>
		<AuthorityRef ref="auth1" />
		<members>
			<LineRef ref="line1" />
			<LineRef ref="line2" />
		</members>
	</Network>
</root>`

	err := loader.parseNetworksFromXML([]byte(xmlData), repo)
	if err != nil {
		t.Fatalf("parseNetworksFromXML() failed: %v", err)
	}

	// Check that network was saved (we need to add Network support to mock repo)
	if len(repo.entities) == 0 {
		t.Error("Expected network to be saved")
	}
}

func TestDefaultNetexDatasetLoader_RepositoryError(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{saveErr: bytes.ErrTooLarge}

	frame := &model.ResourceFrame{
		Authorities: &model.Authorities{
			Authority: []model.Authority{
				{ID: "auth1", Name: "Authority 1"},
			},
		},
	}

	err := loader.loadResourceFrame(frame, repo)
	if err == nil {
		t.Error("Expected error when repository.SaveEntity fails")
	}

	if !strings.Contains(err.Error(), "failed to save authority") {
		t.Errorf("Error message should mention failed to save authority, got: %v", err)
	}
}

func TestDefaultNetexDatasetLoader_EmptyFrames(t *testing.T) {
	loader := &DefaultNetexDatasetLoader{}
	repo := &mockNetexRepository{}

	// Test service frame with empty collections
	frame := &model.ServiceFrame{
		Lines:           &model.Lines{},
		Routes:          &model.Routes{},
		JourneyPatterns: &model.JourneyPatterns{},
	}

	err := loader.loadServiceFrame(frame, repo)
	if err != nil {
		t.Errorf("loadServiceFrame() with empty collections should not error, got: %v", err)
	}

	// Verify no entities were saved
	if len(repo.entities) != 0 {
		t.Errorf("Expected 0 entities saved, got %d", len(repo.entities))
	}
}
