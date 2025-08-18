package loader

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strings"
	"sync"

	"github.com/theoremus-urban-solutions/netex-gtfs-converter/model"
	"github.com/theoremus-urban-solutions/netex-gtfs-converter/producer"
)

// StreamingNetexDatasetLoader implements large file streaming with memory optimization
type StreamingNetexDatasetLoader struct {
	maxMemoryMB      int
	concurrentFiles  int
	bufferSize       int
	progressCallback func(filename string, processed, total int64)
}

// NewStreamingNetexDatasetLoader creates a new streaming loader optimized for large datasets
func NewStreamingNetexDatasetLoader() producer.NetexDatasetLoader {
	return &StreamingNetexDatasetLoader{
		maxMemoryMB:     512, // Default 512MB memory limit
		concurrentFiles: runtime.NumCPU(),
		bufferSize:      64 * 1024, // 64KB buffer
	}
}

// SetMemoryLimit sets the maximum memory usage in MB
func (l *StreamingNetexDatasetLoader) SetMemoryLimit(memoryMB int) {
	l.maxMemoryMB = memoryMB
}

// SetConcurrency sets the number of concurrent file processors
func (l *StreamingNetexDatasetLoader) SetConcurrency(concurrent int) {
	l.concurrentFiles = concurrent
}

// SetProgressCallback sets a callback for progress reporting
func (l *StreamingNetexDatasetLoader) SetProgressCallback(callback func(filename string, processed, total int64)) {
	l.progressCallback = callback
}

// Load implements NetexDatasetLoader with streaming and memory optimization
func (l *StreamingNetexDatasetLoader) Load(data io.Reader, repository producer.NetexRepository) error {
	// Read all data into memory first (we need it for ZIP processing)
	zipData, err := ioutil.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read ZIP data: %w", err)
	}

	// Determine if it's a ZIP file
	if len(zipData) >= 4 && zipData[0] == 'P' && zipData[1] == 'K' {
		return l.loadFromZIPStreaming(zipData, repository)
	}

	// Single XML file
	return l.loadFromXMLStreaming(bytes.NewReader(zipData), repository, "input.xml")
}

// loadFromZIPStreaming processes ZIP files with controlled memory usage
func (l *StreamingNetexDatasetLoader) loadFromZIPStreaming(zipData []byte, repository producer.NetexRepository) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to open ZIP archive: %w", err)
	}

	// Filter XML files and sort by size (process smaller files first)
	xmlFiles := make([]*zip.File, 0)
	totalSize := int64(0)

	for _, file := range zipReader.File {
		if l.isXMLFile(file.Name) {
			xmlFiles = append(xmlFiles, file)
			totalSize += int64(file.UncompressedSize64)
		}
	}

	if len(xmlFiles) == 0 {
		return fmt.Errorf("no XML files found in archive")
	}

	// Process files with controlled concurrency
	semaphore := make(chan struct{}, l.concurrentFiles)
	var wg sync.WaitGroup
	var processingError error
	var errorMutex sync.Mutex

	processedSize := int64(0)
	var sizeMutex sync.Mutex

	for _, file := range xmlFiles {
		wg.Add(1)
		go func(f *zip.File) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if err := l.processZIPFile(f, repository); err != nil {
				errorMutex.Lock()
				if processingError == nil {
					processingError = fmt.Errorf("failed to process file %s: %w", f.Name, err)
				}
				errorMutex.Unlock()
				return
			}

			// Update progress
			sizeMutex.Lock()
			processedSize += int64(f.UncompressedSize64)
			if l.progressCallback != nil {
				l.progressCallback(f.Name, processedSize, totalSize)
			}
			sizeMutex.Unlock()
		}(file)
	}

	wg.Wait()
	return processingError
}

// processZIPFile processes a single file from ZIP archive
func (l *StreamingNetexDatasetLoader) processZIPFile(file *zip.File, repository producer.NetexRepository) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	return l.loadFromXMLStreaming(rc, repository, file.Name)
}

// loadFromXMLStreaming processes XML with streaming parser
func (l *StreamingNetexDatasetLoader) loadFromXMLStreaming(reader io.Reader, repository producer.NetexRepository, filename string) error {
	// Use buffered reader for better performance
	bufferedReader := bufio.NewReaderSize(reader, l.bufferSize)

	// Create streaming XML decoder
	decoder := xml.NewDecoder(bufferedReader)

	// Track processing context
	ctx := &streamingContext{
		filename:   filename,
		repository: repository,
		processed:  0,
	}

	// Process XML tokens in streaming fashion
	return l.processXMLStream(decoder, ctx)
}

// streamingContext holds context for streaming processing
type streamingContext struct {
	filename   string
	repository producer.NetexRepository
	processed  int64
	inFrame    string
	depth      int
}

// processXMLStream processes XML tokens in a streaming manner
func (l *StreamingNetexDatasetLoader) processXMLStream(decoder *xml.Decoder, ctx *streamingContext) error {
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("XML parsing error at position %d: %w", ctx.processed, err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			ctx.depth++
			if err := l.handleStartElement(decoder, &t, ctx); err != nil {
				return err
			}
		case xml.EndElement:
			ctx.depth--
			l.handleEndElement(&t, ctx)
		}

		ctx.processed++
	}

	return nil
}

// handleStartElement processes XML start elements
func (l *StreamingNetexDatasetLoader) handleStartElement(decoder *xml.Decoder, element *xml.StartElement, ctx *streamingContext) error {
	switch element.Name.Local {
	case "ResourceFrame", "ServiceFrame", "ServiceCalendarFrame", "TimetableFrame", "SiteFrame", "GeneralFrame":
		ctx.inFrame = element.Name.Local
		return nil

	// High-level entities that should be processed immediately
	case "Authority":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.Authority{} })
	case "Network":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.Network{} })
	case "Line":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.Line{} })
	case "Route":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.Route{} })
	case "JourneyPattern":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.JourneyPattern{} })
	case "ServiceJourneyPattern":
		return l.processServiceJourneyPattern(decoder, element, ctx)
	case "DestinationDisplay":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.DestinationDisplay{} })
	case "ScheduledStopPoint":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.ScheduledStopPoint{} })
	case "StopPointInJourneyPattern":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.StopPointInJourneyPattern{} })
	case "ServiceJourney":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.ServiceJourney{} })
	case "DatedServiceJourney":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.DatedServiceJourney{} })
	case "ServiceJourneyInterchange":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.ServiceJourneyInterchange{} })
	case "DayType":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.DayType{} })
	case "OperatingDay":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.OperatingDay{} })
	case "OperatingPeriod":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.OperatingPeriod{} })
	case "DayTypeAssignment":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.DayTypeAssignment{} })
	case "StopPlace":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.StopPlace{} })
	case "Quay":
		return l.processEntity(decoder, element, ctx, func() interface{} { return &model.Quay{} })
	}

	return nil
}

// handleEndElement processes XML end elements
func (l *StreamingNetexDatasetLoader) handleEndElement(element *xml.EndElement, ctx *streamingContext) {
	switch element.Name.Local {
	case "ResourceFrame", "ServiceFrame", "ServiceCalendarFrame", "TimetableFrame", "SiteFrame":
		ctx.inFrame = ""
	}
}

// processEntity processes a single XML entity
func (l *StreamingNetexDatasetLoader) processEntity(decoder *xml.Decoder, element *xml.StartElement, ctx *streamingContext, factory func() interface{}) error {
	entity := factory()

	if err := decoder.DecodeElement(entity, element); err != nil {
		return fmt.Errorf("failed to decode %s in %s: %w", element.Name.Local, ctx.filename, err)
	}

	// Save entity to repository
	if err := ctx.repository.SaveEntity(entity); err != nil {
		return fmt.Errorf("failed to save %s in %s: %w", element.Name.Local, ctx.filename, err)
	}

	return nil
}

// processServiceJourneyPattern handles ServiceJourneyPattern elements specially
func (l *StreamingNetexDatasetLoader) processServiceJourneyPattern(decoder *xml.Decoder, element *xml.StartElement, ctx *streamingContext) error {
	// Parse as ServiceJourneyPattern first
	var sjp model.ServiceJourneyPattern
	if err := decoder.DecodeElement(&sjp, element); err != nil {
		return fmt.Errorf("failed to decode ServiceJourneyPattern in %s: %w", ctx.filename, err)
	}

	// Convert to JourneyPattern and save
	jp := sjp.ToJourneyPattern()
	if err := ctx.repository.SaveEntity(jp); err != nil {
		return fmt.Errorf("failed to save journey pattern in %s: %w", ctx.filename, err)
	}

	return nil
}

// isXMLFile checks if a file is an XML file
func (l *StreamingNetexDatasetLoader) isXMLFile(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.HasSuffix(lower, ".xml")
}

// MemoryStats provides memory usage statistics
type MemoryStats struct {
	HeapAlloc    uint64
	HeapInuse    uint64
	HeapIdle     uint64
	HeapReleased uint64
	GCCycles     uint32
	LastGCTime   uint64
}

// GetMemoryStats returns current memory usage statistics
func (l *StreamingNetexDatasetLoader) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		HeapAlloc:    m.HeapAlloc,
		HeapInuse:    m.HeapInuse,
		HeapIdle:     m.HeapIdle,
		HeapReleased: m.HeapReleased,
		GCCycles:     m.NumGC,
		LastGCTime:   m.LastGC,
	}
}

// ForceGC triggers garbage collection if memory usage is high
func (l *StreamingNetexDatasetLoader) ForceGC() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Force GC if we're using more than the configured limit
	if m.HeapInuse > uint64(l.maxMemoryMB)*1024*1024 {
		runtime.GC()
	}
}
