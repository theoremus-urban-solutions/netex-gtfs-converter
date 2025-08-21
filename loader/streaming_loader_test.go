package loader

import (
	"io"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	testXMLData = `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<Authority id="auth1" version="1">
		<Name>Authority 1</Name>
	</Authority>
</root>`
)

func TestNewStreamingNetexDatasetLoader(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader()
	if loader == nil {
		t.Fatal("NewStreamingNetexDatasetLoader() returned nil")
	}

	streamingLoader, ok := loader.(*StreamingNetexDatasetLoader)
	if !ok {
		t.Fatal("NewStreamingNetexDatasetLoader() did not return *StreamingNetexDatasetLoader")
	}

	// Check default values
	if streamingLoader.maxMemoryMB != 512 {
		t.Errorf("Expected default maxMemoryMB 512, got %d", streamingLoader.maxMemoryMB)
	}

	if streamingLoader.concurrentFiles != runtime.NumCPU() {
		t.Errorf("Expected default concurrentFiles %d, got %d", runtime.NumCPU(), streamingLoader.concurrentFiles)
	}

	if streamingLoader.bufferSize != 64*1024 {
		t.Errorf("Expected default bufferSize %d, got %d", 64*1024, streamingLoader.bufferSize)
	}
}

func TestStreamingNetexDatasetLoader_Configuration(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader().(*StreamingNetexDatasetLoader)

	// Test SetMemoryLimit
	loader.SetMemoryLimit(1024)
	if loader.maxMemoryMB != 1024 {
		t.Errorf("Expected maxMemoryMB 1024, got %d", loader.maxMemoryMB)
	}

	// Test SetConcurrency
	loader.SetConcurrency(4)
	if loader.concurrentFiles != 4 {
		t.Errorf("Expected concurrentFiles 4, got %d", loader.concurrentFiles)
	}

	// Test SetProgressCallback
	callbackCalled := false
	loader.SetProgressCallback(func(filename string, processed, total int64) {
		callbackCalled = true
	})

	if loader.progressCallback == nil {
		t.Error("Expected progressCallback to be set")
	}

	// Trigger callback to test
	loader.progressCallback("test.xml", 100, 200)
	if !callbackCalled {
		t.Error("Progress callback was not called")
	}
}

func TestStreamingNetexDatasetLoader_LoadSingleXML(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader()
	repo := &mockNetexRepository{}

	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<PublicationDelivery xmlns="http://www.netex.org.uk/netex">
	<ResourceFrame>
		<Authority id="test-authority" version="1">
			<Name>Test Authority</Name>
		</Authority>
	</ResourceFrame>
</PublicationDelivery>`

	reader := strings.NewReader(xmlData)
	err := loader.Load(reader, repo)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify authority was loaded
	authorities := repo.GetAuthorities()
	if len(authorities) != 1 {
		t.Errorf("Expected 1 authority, got %d", len(authorities))
	}

	if len(authorities) > 0 && authorities[0].ID != "test-authority" {
		t.Errorf("Expected authority ID 'test-authority', got '%s'", authorities[0].ID)
	}
}

func TestStreamingNetexDatasetLoader_LoadInvalidData(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader()
	repo := &mockNetexRepository{}

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "empty input",
			input:     "",
			expectErr: false, // Empty XML should not error, just no entities
		},
		{
			name:      "invalid XML",
			input:     "<?xml version=\"1.0\"?><root><unclosed>",
			expectErr: true,
		},
		{
			name:      "corrupted zip header",
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

func TestStreamingNetexDatasetLoader_isXMLFile(t *testing.T) {
	loader := &StreamingNetexDatasetLoader{}

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.xml", true},
		{"TEST.XML", true},
		{"data.XML", true},
		{"file.txt", false},
		{"readme.md", false},
		{"", false},
		{"xml", false},
		{"test.xml.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := loader.isXMLFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isXMLFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestStreamingNetexDatasetLoader_MemoryStats(t *testing.T) {
	loader := &StreamingNetexDatasetLoader{}

	stats := loader.GetMemoryStats()

	// Basic checks that stats are populated
	if stats.HeapAlloc == 0 {
		t.Error("Expected HeapAlloc to be non-zero")
	}
	if stats.HeapInuse == 0 {
		t.Error("Expected HeapInuse to be non-zero")
	}
}

func TestStreamingNetexDatasetLoader_ForceGC(t *testing.T) {
	loader := &StreamingNetexDatasetLoader{
		maxMemoryMB: 1, // Very low limit to trigger GC
	}

	// Get initial GC count
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)
	initialGCCount := initialStats.NumGC

	// Force GC
	loader.ForceGC()

	// Give some time for GC to potentially run
	time.Sleep(10 * time.Millisecond)

	// Check if GC was triggered (this is not guaranteed, but we can try)
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// ForceGC should not panic and should complete successfully
	// We can't reliably test if GC actually ran due to Go's GC behavior
	if finalStats.NumGC < initialGCCount {
		t.Error("GC count decreased, which should not happen")
	}
}

func TestStreamingNetexDatasetLoader_ProcessEntities(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader()
	repo := &mockNetexRepository{}

	// Test XML with multiple entity types
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<Authority id="auth1" version="1">
		<Name>Authority 1</Name>
	</Authority>
	<Network id="net1" version="1">
		<Name>Network 1</Name>
		<AuthorityRef ref="auth1" />
	</Network>
	<Line id="line1" version="1">
		<Name>Line 1</Name>
		<AuthorityRef>auth1</AuthorityRef>
	</Line>
	<Route id="route1" version="1">
		<Name>Route 1</Name>
		<LineRef ref="line1" />
	</Route>
	<StopPlace id="stop1" version="1">
		<Name>Stop Place 1</Name>
		<Centroid>
			<Location>
				<Longitude>10.0</Longitude>
				<Latitude>60.0</Latitude>
			</Location>
		</Centroid>
	</StopPlace>
	<Quay id="quay1" version="1">
		<Name>Platform 1</Name>
		<Centroid>
			<Location>
				<Longitude>10.0</Longitude>
				<Latitude>60.0</Latitude>
			</Location>
		</Centroid>
	</Quay>
</root>`

	reader := strings.NewReader(xmlData)
	err := loader.Load(reader, repo)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify all entities were loaded
	if len(repo.entities) < 5 { // Should have at least Authority, Network, Line, Route, StopPlace, Quay
		t.Errorf("Expected at least 5 entities loaded, got %d", len(repo.entities))
	}
}

func TestStreamingNetexDatasetLoader_ProgressCallback(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader().(*StreamingNetexDatasetLoader)
	repo := &mockNetexRepository{}

	// Track progress callback calls
	var callbackCalls []struct {
		filename  string
		processed int64
		total     int64
	}

	loader.SetProgressCallback(func(filename string, processed, total int64) {
		callbackCalls = append(callbackCalls, struct {
			filename  string
			processed int64
			total     int64
		}{filename, processed, total})
	})

	reader := strings.NewReader(testXMLData)
	err := loader.Load(reader, repo)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// For single XML files, we don't expect progress callbacks since it's processed directly
	// This mainly tests that the callback doesn't cause issues
}

func TestStreamingNetexDatasetLoader_ConcurrentProcessing(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader().(*StreamingNetexDatasetLoader)
	repo := &mockNetexRepository{}

	// Set low concurrency for predictable testing
	loader.SetConcurrency(2)

	if loader.concurrentFiles != 2 {
		t.Errorf("Expected concurrentFiles 2, got %d", loader.concurrentFiles)
	}

	// Test with simple XML
	reader := strings.NewReader(testXMLData)
	err := loader.Load(reader, repo)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
}

func TestStreamingNetexDatasetLoader_LargeData(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader()
	repo := &mockNetexRepository{}

	// Create XML with many entities to test streaming behavior
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?><root>`)

	// Generate 100 authorities
	for i := 0; i < 100; i++ {
		xmlBuilder.WriteString(`<Authority id="auth`)
		xmlBuilder.WriteString(strings.Repeat("0", 10)) // Add padding to make it larger
		xmlBuilder.WriteString(`" version="1"><Name>Authority </Name></Authority>`)
	}

	xmlBuilder.WriteString(`</root>`)
	xmlData := xmlBuilder.String()

	reader := strings.NewReader(xmlData)
	err := loader.Load(reader, repo)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	authorities := repo.GetAuthorities()
	if len(authorities) != 100 {
		t.Errorf("Expected 100 authorities, got %d", len(authorities))
	}
}

func TestStreamingNetexDatasetLoader_ErrorHandling(t *testing.T) {
	loader := NewStreamingNetexDatasetLoader()
	repo := &mockNetexRepository{saveErr: io.ErrUnexpectedEOF}

	reader := strings.NewReader(testXMLData)
	err := loader.Load(reader, repo)
	if err == nil {
		t.Error("Expected error when repository SaveEntity fails")
	}

	if !strings.Contains(err.Error(), "failed to save") {
		t.Errorf("Error should mention 'failed to save', got: %v", err)
	}
}
