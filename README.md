# NeTEx to GTFS Converter (Go)

A high-performance Go implementation of the NeTEx to GTFS converter, providing efficient conversion of Network Timetable Exchange datasets into GTFS (General Transit Feed Specification) format.

## Overview

This tool converts NeTEx (Network Timetable Exchange) datasets into GTFS format with focus on European NeTEx profiles. It features comprehensive error handling, memory optimization, and extensible architecture.

## Features

- **European NeTEx Profile Support**: Specialized handling of European transit data standards
- **Performance Optimized**: Memory-efficient processing with streaming capabilities  
- **Extensible Architecture**: Producer pattern for easy customization and extension
- **Comprehensive Route Type Mapping**: Full support for GTFS extended route types
- **Robust Error Handling**: Recovery mechanisms and detailed validation reporting
- **CLI Interface**: Feature-rich command-line tool with multiple output options
- **Advanced Stop Time Production**: Enhanced algorithms for accurate timetable conversion
- **Calendar Management**: Sophisticated service calendar generation and holiday detection
- **Geometry Processing**: Shape generation and spatial data handling

## Installation

### Prerequisites

- Go 1.21 or later

### Quick Install

```bash
# Clone the repository
git clone https://github.com/theoremus-urban-solutions/netex-gtfs-converter
cd netex-gtfs-converter

# Build using Make
make build

# Or install directly
make install
```

### Development Setup

```bash
# Install development tools
make dev-tools

# Run development checks
make dev
```

## Usage

### Basic Usage

```bash
# Using the built binary
./bin/netex-gtfs-converter --codespace FR --netex data.zip --output gtfs.zip

# Convert only stops
./bin/netex-gtfs-converter --stops stops.zip --stops-only --output stops-only.zip

# Example with French data
./bin/netex-gtfs-converter --netex fluo-grand-est-riv-netex.zip --codespace FR --output /tmp/gtfs.zip

# Using installed binary
netex-gtfs-converter --help
```

### Command Line Options

| Option | Description | Required |
|--------|-------------|----------|
| `--codespace` | NeTEx codespace | For timetable conversion |
| `--netex` | NeTEx timetable ZIP file | For timetable conversion |
| `--stops` | NeTEx stops ZIP file | Optional |
| `--output` | Output GTFS ZIP file | No (default: gtfs.zip) |
| `--stops-only` | Convert only stops | No |
| `--verbose` | Enable verbose logging | No |
| `--help` | Show help message | No |

### Output Features

The converter automatically ensures complete GTFS compliance by:
- ✅ Generating all required GTFS files (`agency.txt`, `routes.txt`, `stops.txt`, `trips.txt`, `stop_times.txt`, `calendar.txt`, `feed_info.txt`)
- ✅ Fixing CSV header naming issues (e.g., `feed_publisher_url`)
- ✅ Creating basic service calendars and trip schedules when data is incomplete
- ✅ Providing error recovery and validation reporting

### Profile Types

- **european**: European NeTEx Profile (flexible requirements)

## Architecture

The converter follows a producer pattern architecture similar to the Java version:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   NeTEx Data    │───▶│   Repositories  │───▶│   Producers     │
│   (ZIP/XML)     │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                              ┌─────────────────┐
                                              │   GTFS Data     │
                                              │   (CSV/ZIP)     │
                                              └─────────────────┘
```

### Core Components

- **Exporter**: Main conversion orchestrator
- **Producers**: Convert specific NeTEx entities to GTFS entities
- **Repositories**: In-memory data storage and access
- **Loaders**: Parse NeTEx ZIP archives and XML data
- **Serializers**: Generate GTFS CSV files and ZIP archives

### Producer Interfaces

- `AgencyProducer`: Converts NeTEx Authority to GTFS Agency
- `RouteProducer`: Converts NeTEx Line to GTFS Route
- `TripProducer`: Converts NeTEx ServiceJourney to GTFS Trip
- `StopProducer`: Converts NeTEx Quay to GTFS Stop
- `StopTimeProducer`: Converts NeTEx TimetabledPassingTime to GTFS StopTime
- `ServiceCalendarProducer`: Converts service patterns to GTFS Calendar
- `ShapeProducer`: Converts route geometry to GTFS Shape
- `TransferProducer`: Converts interchanges to GTFS Transfers

## Profile Configuration

The converter supports different NeTEx profiles through configuration:

### European Profile
- Quays may have their own names
- Quays may not inherit transport mode from parent StopPlace
- StopPlace names are optional
- DestinationDisplay is optional

## Route Type Mapping

The converter includes comprehensive mapping from NeTEx transport modes/submodes to GTFS extended route types:

| NeTEx Mode | NeTEx Submode | GTFS Route Type | Code |
|------------|---------------|-----------------|------|
| bus | expressBus | Express Bus Service | 702 |
| bus | localBus | Local Bus Service | 704 |
| rail | longDistance | Long Distance Trains | 102 |
| rail | highSpeed | High Speed Rail Service | 101 |
| tram | cityTram | City Tram Service | 901 |
| ferry | | Ferry Service | 1200 |

See the complete mapping in `internal/model/route_types.go`.

## Extension Points

The converter is designed for extensibility. You can customize the conversion process by implementing custom producers:

```go
// Custom agency producer
type CustomAgencyProducer struct {
    producer.DefaultAgencyProducer
}

func (p *CustomAgencyProducer) Produce(authority *model.Authority) (*model.Agency, error) {
    // Custom logic here
    return p.DefaultAgencyProducer.Produce(authority)
}

// Use custom producer
exporter := exporter.NewDefaultGtfsExporter(codespace, stopAreaRepo, profileConfig)
exporter.SetAgencyProducer(&CustomAgencyProducer{})
```

## Development

### Project Structure

```
github.com/theoremus-urban-solutions/netex-gtfs-converter/
├── cmd/
│   └── converter/
│       ├── main.go              # CLI entry point
│       └── main_test.go         # CLI tests
├── benchmark/                   # Performance benchmarks
├── calendar/                    # Calendar and service management
├── config/                      # Configuration handling
├── exporter/                    # GTFS export functionality
├── geometry/                    # Spatial processing
├── loader/                      # NeTEx data loading
├── memory/                      # Memory optimization
├── model/                       # Data models and structures
├── producer/                    # Data transformation producers  
├── repository/                  # Data access layer
├── validation/                  # Data validation
├── testdata/                    # Test data files
├── docs/                        # Generated documentation
├── examples/                    # Usage examples
├── go.mod
├── Makefile                     # Build automation
└── README.md
```

### Available Make Targets

```bash
# Development
make dev-tools          # Install development tools
make build              # Build the CLI binary
make install            # Install the CLI binary
make dev                # Quick development check

# Testing & Quality
make test               # Run all tests
make test-verbose       # Run tests with verbose output
make benchmark          # Run performance benchmarks
make coverage           # Generate test coverage report
make lint               # Run linter
make validate           # Run all validation checks

# Build & Release
make build-all          # Build for all platforms
make release-prep       # Prepare for release
make clean              # Clean build artifacts

# Documentation
make docs               # Generate documentation
make docs-serve         # Start documentation server
make stats              # Show project statistics
```

### Building

```bash
# Quick build
make build

# Build for all platforms
make build-all

# Manual build
go build -o bin/netex-gtfs-converter ./cmd/converter
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run benchmarks
make benchmark

# Run specific package tests
go test ./calendar
go test ./producer
```

## Performance & Optimization

This Go implementation provides significant performance improvements:

- **Memory Efficiency**: Advanced memory optimization with streaming processing
- **Concurrent Processing**: Parallel data transformation and validation
- **Single Binary Deployment**: No runtime dependencies or JVM required
- **Comprehensive Benchmarking**: Built-in performance monitoring and profiling
- **Optimized Data Structures**: Custom repositories for efficient data access

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Install development tools (`make dev-tools`)
4. Make your changes following the existing code style
5. Add tests for new functionality
6. Run validation checks (`make validate`)
7. Run benchmarks to ensure performance (`make benchmark`)
8. Commit your changes (`git commit -m 'Add amazing feature'`)
9. Push to the branch (`git push origin feature/amazing-feature`)
10. Submit a pull request

### Development Workflow

```bash
# Setup development environment
make dev-tools

# Make changes and test
make dev

# Full validation before PR
make check
```

## Documentation

Generate and view comprehensive documentation:

```bash
# Generate API documentation
make docs

# Serve documentation locally
make docs-serve
# Visit: http://localhost:6060/pkg/github.com/theoremus-urban-solutions/netex-gtfs-converter/

# View project statistics
make stats
```

Generated documentation includes:
- API reference for all packages
- Package documentation  
- HTML documentation (if godoc is available)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes, new features, and bug fixes.

## License

This project is licensed under the same terms as the original Java implementation.

## Acknowledgments

This Go implementation builds upon the excellent foundation provided by the [Entur team](https://github.com/entur/netex-gtfs-converter-java) and their Java NeTEx to GTFS converter. We extend our gratitude for their pioneering work in transit data standardization.
