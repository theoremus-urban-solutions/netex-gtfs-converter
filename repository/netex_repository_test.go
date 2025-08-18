package repository

import (
	"testing"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
)

func TestDefaultNetexRepository(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Test saving and retrieving a line
	line := &model.Line{
		ID:           "test-line",
		Name:         "Test Line",
		ShortName:    "TL",
		AuthorityRef: "test-authority",
	}

	err := repo.SaveEntity(line)
	if err != nil {
		t.Fatalf("SaveEntity() failed: %v", err)
	}

	lines := repo.GetLines()
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	if lines[0].ID != line.ID {
		t.Errorf("Expected line ID %s, got %s", line.ID, lines[0].ID)
	}

	// Test saving and retrieving an authority
	authority := &model.Authority{
		ID:   "test-authority",
		Name: "Test Authority",
	}

	err = repo.SaveEntity(authority)
	if err != nil {
		t.Fatalf("SaveEntity() failed: %v", err)
	}

	retrievedAuthority := repo.GetAuthorityById("test-authority")
	if retrievedAuthority == nil {
		t.Fatal("GetAuthorityById() returned nil")
	}

	if retrievedAuthority.Name != authority.Name {
		t.Errorf("Expected authority name %s, got %s", authority.Name, retrievedAuthority.Name)
	}

	// Test route association with line
	route := &model.Route{
		ID:      "test-route",
		LineRef: "test-line",
		Name:    "Test Route",
	}

	err = repo.SaveEntity(route)
	if err != nil {
		t.Fatalf("SaveEntity() failed: %v", err)
	}

	routes := repo.GetRoutesByLine(line)
	if len(routes) != 1 {
		t.Fatalf("Expected 1 route for line, got %d", len(routes))
	}

	if routes[0].ID != route.ID {
		t.Errorf("Expected route ID %s, got %s", route.ID, routes[0].ID)
	}
}

func TestDefaultNetexRepository_GetAuthorityIdForLine(t *testing.T) {
	repo := NewDefaultNetexRepository()

	line := &model.Line{
		ID:           "test-line",
		AuthorityRef: "test-authority",
	}

	authorityId := repo.GetAuthorityIdForLine(line)
	if authorityId != "test-authority" {
		t.Errorf("Expected authority ID %s, got %s", "test-authority", authorityId)
	}
}

func TestDefaultNetexRepository_GetTimeZone(t *testing.T) {
	repo := NewDefaultNetexRepository()

	// Should return default timezone
	tz := repo.GetTimeZone()
	if tz == "" {
		t.Error("GetTimeZone() returned empty string")
	}

	// Default should be Europe/Oslo
	if tz != "Europe/Oslo" {
		t.Errorf("Expected default timezone 'Europe/Oslo', got '%s'", tz)
	}
}
