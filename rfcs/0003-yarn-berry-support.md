# RFC 0003: Add Yarn Berry Support

## Summary

This RFC proposes adding support for Yarn Berry (v2+) alongside the existing Yarn Classic (v1.x) support in the Paketo Yarn Start buildpack. The buildpack will automatically detect which version of Yarn is being used and adapt accordingly.

## Motivation

Yarn Berry was released as a major evolution of the Yarn package manager with significant changes:

- **Different lock file format**: Berry uses YAML format vs Classic's custom format
- **Plug'n'Play (PnP) by default**: Different dependency resolution strategy
- **Configuration changes**: Uses `.yarnrc.yml` instead of `.yarnrc`
- **Version pinning**: Projects can pin specific Yarn versions via `packageManager` field

Many teams are migrating to Yarn Berry for its improved performance, better workspace support, and modern features. The buildpack should support both versions seamlessly.

## Detailed Design

### Version Detection Strategy

The buildpack will detect Yarn Berry vs Classic using this precedence order:

1. **`.yarnrc.yml` file exists** → Yarn Berry
2. **`packageManager` field in `package.json`** → Check version (e.g., `yarn@3.6.0` = Berry)
3. **`yarn.lock` format** → YAML format = Berry, custom format = Classic  
4. **Default to Classic** → For backward compatibility

### Implementation Details

#### New Components

1. **`YarnDetector`** - Handles version detection logic
2. **`YarnVersion` enum** - Represents Classic vs Berry
3. **Updated detection metadata** - Includes `yarn-version` field for upstream buildpacks

#### Detection Logic

```go
type YarnDetector struct {
    projectPath string
}

func (d *YarnDetector) DetectYarnVersion() (YarnVersion, error) {
    // 1. Check for .yarnrc.yml (strongest Berry indicator)
    // 2. Check packageManager field in package.json  
    // 3. Check yarn.lock format (YAML vs custom)
    // 4. Default to Classic
}
```

#### Build Plan Changes

The buildpack will add yarn version metadata to build requirements:

```go
{
    Name: "yarn",
    Metadata: map[string]interface{}{
        "launch":       true,
        "yarn-version": "Berry", // or "Classic"
    },
}
```

### Script Execution

Both Yarn versions use identical script execution (e.g., `yarn start`), so no changes are needed in the build logic beyond detection and logging.

### Backward Compatibility

- **Existing Yarn Classic projects**: Continue working unchanged
- **Default behavior**: Still defaults to Classic when no Berry indicators found
- **Build plan metadata**: New `yarn-version` field is optional for consuming buildpacks

## Testing Strategy

### Unit Tests
- **YarnDetector**: Test all detection scenarios and precedence rules
- **Integration**: Test with both Classic and Berry test applications
- **Regression**: Ensure existing Classic functionality unchanged

### Integration Tests
- **Yarn Berry app**: New test data with `.yarnrc.yml`, Berry lock file, `packageManager` field
- **Live reload**: Test Berry apps with `BP_LIVE_RELOAD_ENABLED=true`
- **Edge cases**: Malformed files, mixed indicators

## Documentation

### README Updates
- Document automatic version detection
- List detection criteria and precedence
- Clarify that both versions are supported
- Note that no user configuration is needed

### Examples
- Show sample Berry project structure
- Demonstrate `packageManager` field usage
- Explain `.yarnrc.yml` configuration

## Alternatives Considered

### Manual Configuration
**Option**: Add environment variable like `BP_YARN_VERSION=berry`
**Rejected**: Auto-detection is more user-friendly and less error-prone

### Separate Buildpacks  
**Option**: Create separate `yarn-berry-start` buildpack
**Rejected**: Single buildpack with auto-detection reduces complexity

### Lock File Only Detection
**Option**: Only use `yarn.lock` format for detection
**Rejected**: `.yarnrc.yml` and `packageManager` are stronger indicators

## Implementation Status

- ✅ **YarnDetector implementation** with comprehensive version detection
- ✅ **Unit tests** covering all detection scenarios and edge cases  
- ✅ **Integration test data** for Yarn Berry applications
- ✅ **Updated documentation** explaining Berry support
- ✅ **Build plan metadata** for upstream buildpack coordination
- ✅ **Backward compatibility** verified with existing tests

## Future Considerations

### Enhanced Berry Features
- **Workspace-specific detection**: Different Berry versions per workspace
- **PnP-specific optimizations**: Leverage PnP for faster builds
- **Corepack integration**: Support Node.js Corepack for version management

### Monitoring
- **Usage metrics**: Track Classic vs Berry adoption
- **Performance comparison**: Measure build time differences
- **Error analysis**: Monitor Berry-specific issues

## Conclusion

This implementation provides seamless Yarn Berry support while maintaining full backward compatibility with Yarn Classic. The automatic detection ensures zero configuration for users while giving upstream buildpacks the information they need to optimize for specific Yarn versions.

The feature is production-ready with comprehensive test coverage and follows Paketo buildpack best practices.
