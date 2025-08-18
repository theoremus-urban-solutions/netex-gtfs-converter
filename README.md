# NeTEx to GTFS Converter (Go)

A Go implementation of the NeTEx to GTFS converter, based on the [Java version](https://github.com/entur/netex-gtfs-converter-java).

## Overview

This tool converts NeTEx (Network Timetable Exchange) datasets into GTFS (General Transit Feed Specification) format. It supports only European NeTEx profile.

## Features

- **Multiple Profile Support**: European
- **Flexible Configuration**: Profile-aware conversion with configurable requirements
- **Extensible Architecture**: Producer pattern for easy customization
- **Comprehensive Route Type Mapping**: Full support for GTFS extended route types
- **CLI Interface**: Easy-to-use command-line tool
- **Stop-Only Conversion**: Convert only stop data without timetable information

## Installation

### Prerequisites

- Go 1.21 or later

### Build from Source

```bash
git clone <repository-url>
cd github.com/theoremus-urban-solutions/netex-gtfs-converter
go build -o converter/converter converter/main.go
```

## Usage

### Basic Usage

```bash
# Convert full timetable data
./converter/converter --codespace FR --netex data.zip --output gtfs.zip

# Convert only stops
./converter/converter --stops stops.zip --stops-only --output stops-only.zip

# Example with French data
./converter/converter --netex fluo-grand-est-riv-netex.zip --codespace FR --output /tmp/gtfs.zip
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
│       └── main.go              # CLI entry point
├── internal/
│   ├── config/
│   │   └── profile_config.go    # Profile configuration
│   ├── exporter/
│   │   ├── gtfs_exporter.go     # Main exporter interface
│   │   └── errors.go            # Error definitions
│   ├── model/
│   │   ├── netex_models.go      # NeTEx data structures
│   │   ├── gtfs_models.go       # GTFS data structures
│   │   └── route_types.go       # Route type mapping
│   ├── producer/
│   │   └── producer.go          # Producer interfaces
│   ├── repository/
│   │   ├── netex_repository.go  # NeTEx data access
│   │   └── gtfs_repository.go   # GTFS data access
│   ├── loader/
│   │   └── netex_loader.go      # NeTEx data loading
│   ├── serializer/
│   │   └── gtfs_serializer.go   # GTFS serialization
│   └── util/
│       └── geometry.go          # Geometry utilities
├── pkg/
│   ├── utils/
│   └── validation/
├── testdata/                    # Test data files
├── go.mod
└── README.md
```

### Building

```bash
# Build the converter
go build -o converter cmd/converter/main.go

# Run tests
go test ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o converter-linux cmd/converter/main.go
GOOS=windows GOARCH=amd64 go build -o converter.exe cmd/converter/main.go
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test ./exporter
```

## Comparison with Java Version

This Go implementation maintains compatibility with the Java version while providing:

- **Better Performance**: Go's efficient memory management and concurrency
- **Simpler Deployment**: Single binary, no JVM required
- **Profile Flexibility**: More configurable profile handling
- **Modern Architecture**: Clean interfaces and dependency injection

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the same license as the original Java version.

## Acknowledgments

This implementation is based on the excellent work of the [Entur team](https://github.com/entur/netex-gtfs-converter-java) and their Java NeTEx to GTFS converter.










