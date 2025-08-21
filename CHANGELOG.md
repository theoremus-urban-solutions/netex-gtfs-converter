# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive test coverage across all modules
- Performance benchmarking suite with memory profiling
- Advanced stop time production algorithms
- Enhanced error recovery mechanisms
- Memory optimization features with streaming processing
- Sophisticated calendar management with holiday detection
- Geometry utilities for shape generation and spatial processing
- European NeTEx profile extensions
- Optimized data repositories for improved performance
- Comprehensive validation framework with detailed reporting
- Integration tests for end-to-end validation
- Edge case handling in producers
- Focused coverage testing for critical components

### Enhanced
- CLI interface with improved argument handling and validation
- Error handling with detailed context and recovery strategies
- Documentation generation with proper API reference
- Build system with comprehensive Make targets
- Project structure reorganization for better maintainability
- Producer pattern implementation for extensibility
- Repository layer with optimized data access patterns
- Export functionality with multiple output formats
- Loader implementations with streaming capabilities

### Fixed
- Documentation generation issues in Makefile
- Memory leaks in large dataset processing
- Route type mapping inconsistencies
- Calendar service edge cases
- Stop area repository performance bottlenecks

### Developer Experience
- Complete Makefile with development, testing, and release targets
- Development tools setup automation
- Code quality checks (linting, formatting, security analysis)
- Automated testing with coverage reporting
- Cross-platform build support
- CI/CD pipeline configuration
- Comprehensive benchmarking suite

## [0.1.0] - Initial Release

### Added
- Initial Go implementation of NeTEx to GTFS converter
- European NeTEx profile support
- Core conversion functionality:
  - Agency conversion from NeTEx Authority
  - Route conversion from NeTEx Line  
  - Trip conversion from NeTEx ServiceJourney
  - Stop conversion from NeTEx Quay
  - Stop time conversion from NeTEx TimetabledPassingTime
  - Service calendar generation
  - Transfer conversion from interchanges
- CLI application with command-line interface
- Basic error handling and validation
- GTFS output with ZIP archive generation
- Route type mapping for European transit modes
- Producer pattern architecture
- Repository-based data access
- XML parsing and data loading capabilities

### Features
- Convert NeTEx datasets to GTFS format
- Support for European NeTEx profile specifications
- Command-line tool for batch processing
- Extensible producer architecture
- Memory-efficient data processing
- GTFS compliance validation
- ZIP archive input/output handling

---

## Guidelines for Future Releases

### Version Numbering
- **Major (X.0.0)**: Breaking API changes, major feature additions
- **Minor (X.Y.0)**: New features, backwards compatible
- **Patch (X.Y.Z)**: Bug fixes, documentation updates

### Change Categories
- **Added**: New features
- **Changed**: Changes in existing functionality  
- **Deprecated**: Soon-to-be removed features
- **Removed**: Features removed in this release
- **Fixed**: Bug fixes
- **Security**: Security vulnerability fixes

### Release Process
1. Update version in relevant files
2. Update CHANGELOG.md with release notes
3. Run `make release-prep` to validate
4. Create release with `make release`
5. Update documentation if needed

### Breaking Changes
Breaking changes should be clearly marked and include:
- Description of the change
- Migration guide for users
- Rationale for the change
- Timeline for deprecation (if applicable)