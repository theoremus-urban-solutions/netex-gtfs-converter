# Pull Request

## Summary

<!-- Brief description of the changes -->

## Type of Change

<!-- Mark the relevant option with an "x" -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)  
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Refactoring (no functional changes, no api changes)
- [ ] ⚡ Performance improvement
- [ ] 🧪 Test improvement
- [ ] 🏗️ Build/CI changes

## Changes Made

<!-- List the specific changes made -->

- 
- 
- 

## Related Issues

<!-- Link to related issues -->

Fixes #(issue number)
Closes #(issue number)
Related to #(issue number)

## Testing

<!-- Describe the tests you've added or run -->

### Test Coverage

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated  
- [ ] CLI tests added/updated (if applicable)
- [ ] Benchmark tests added/updated (if applicable)

### Manual Testing

<!-- Describe any manual testing performed -->

- [ ] Tested with sample GTFS feeds
- [ ] Tested CLI functionality
- [ ] Tested performance impact
- [ ] Tested on multiple platforms

## Documentation

<!-- Mark what documentation was updated -->

- [ ] README updated
- [ ] API documentation updated
- [ ] Examples updated
- [ ] CHANGELOG updated
- [ ] Contributing guidelines updated

## Checklist

<!-- Ensure all items are completed before submitting -->

### Code Quality

- [ ] Code follows the project's style guidelines
- [ ] Self-review of code completed
- [ ] Code is commented, particularly in hard-to-understand areas
- [ ] No new warnings introduced

### Testing & CI

- [ ] All tests pass locally (`go test ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] Code is formatted (`gofmt -s -w .`)
- [ ] CI checks pass

### Compatibility

- [ ] Changes are backward compatible (or breaking changes are documented)
- [ ] Works with Go 1.21+
- [ ] No new external dependencies added (or justified)

## Performance Impact

<!-- If applicable, describe performance implications -->

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance regression (justified and documented)
- [ ] Benchmarks added/updated

## Breaking Changes

<!-- If this PR contains breaking changes, describe them here -->

None

<!-- OR list breaking changes:
- Changed API signature of...
- Removed deprecated function...
- Modified behavior of...
-->

## Screenshots/Output

<!-- If applicable, add screenshots or sample output -->

## Additional Notes

<!-- Add any additional notes, considerations, or questions -->

---

**Reviewer Guidelines:**
- Check that tests adequately cover the changes
- Verify documentation is updated
- Ensure compatibility is maintained
- Test with sample GTFS data if applicable