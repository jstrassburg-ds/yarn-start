package yarnstart

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/fs"
	"gopkg.in/yaml.v2"
)

// YarnVersion represents the version of Yarn being used
type YarnVersion int

const (
	YarnClassic YarnVersion = iota
	YarnBerry
)

// String returns a string representation of the Yarn version
func (v YarnVersion) String() string {
	switch v {
	case YarnClassic:
		return "Classic"
	case YarnBerry:
		return "Berry"
	default:
		return "Unknown"
	}
}

// YarnDetector provides methods to detect Yarn version and configuration
type YarnDetector struct {
	projectPath string
}

// NewYarnDetector creates a new YarnDetector for the given project path
func NewYarnDetector(projectPath string) *YarnDetector {
	return &YarnDetector{
		projectPath: projectPath,
	}
}

// DetectYarnVersion determines which version of Yarn is being used
func (d *YarnDetector) DetectYarnVersion() (YarnVersion, error) {
	// Check for Yarn Berry indicators in order of precedence
	
	// 1. Check for .yarnrc.yml file (strongest Berry indicator)
	yarnrcPath := filepath.Join(d.projectPath, ".yarnrc.yml")
	if exists, err := fs.Exists(yarnrcPath); err != nil {
		return YarnClassic, fmt.Errorf("failed to check for .yarnrc.yml: %w", err)
	} else if exists {
		return YarnBerry, nil
	}

	// 2. Check packageManager field in package.json
	packageJsonPath := filepath.Join(d.projectPath, "package.json")
	if exists, err := fs.Exists(packageJsonPath); err != nil {
		return YarnClassic, fmt.Errorf("failed to check for package.json: %w", err)
	} else if exists {
		if isBerry, err := d.checkPackageManagerField(packageJsonPath); err != nil {
			return YarnClassic, fmt.Errorf("failed to check packageManager field: %w", err)
		} else if isBerry {
			return YarnBerry, nil
		}
	}

	// 3. Check yarn.lock format
	yarnLockPath := filepath.Join(d.projectPath, "yarn.lock")
	if exists, err := fs.Exists(yarnLockPath); err != nil {
		return YarnClassic, fmt.Errorf("failed to check for yarn.lock: %w", err)
	} else if exists {
		if isBerry, err := d.checkLockFileFormat(yarnLockPath); err != nil {
			return YarnClassic, fmt.Errorf("failed to check yarn.lock format: %w", err)
		} else if isBerry {
			return YarnBerry, nil
		}
	}

	// Default to Classic if no Berry indicators found
	return YarnClassic, nil
}

// checkPackageManagerField checks if package.json specifies Yarn 2+ in packageManager field
func (d *YarnDetector) checkPackageManagerField(packageJsonPath string) (bool, error) {
	content, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return false, err
	}

	var pkg struct {
		PackageManager string `json:"packageManager"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		// If JSON is malformed, we can't determine from this field
		// Let the main package parsing handle the error
		return false, nil
	}

	if pkg.PackageManager == "" {
		return false, nil
	}

	// Check if packageManager specifies yarn 2+ (Berry)
	// Format is typically "yarn@3.6.0" or "yarn@berry"
	if strings.HasPrefix(pkg.PackageManager, "yarn@") {
		version := strings.TrimPrefix(pkg.PackageManager, "yarn@")
		if version == "berry" || version == "stable" {
			return true, nil
		}
		
		// Check if version starts with 2, 3, 4, etc. (Berry versions)
		if len(version) > 0 && (version[0] >= '2' && version[0] <= '9') {
			return true, nil
		}
	}

	return false, nil
}

// checkLockFileFormat determines if yarn.lock uses Berry format (YAML) or Classic format
func (d *YarnDetector) checkLockFileFormat(yarnLockPath string) (bool, error) {
	file, err := os.Open(yarnLockPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first few lines to check format
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		// Empty file or read error - default to Classic
		return false, nil
	}

	content := string(buffer[:n])
	
	// Classic yarn.lock starts with "# yarn lockfile v1"
	if strings.Contains(content, "# yarn lockfile v1") {
		return false, nil
	}

	// Berry lockfile typically contains "__metadata:" or YAML structure
	if strings.Contains(content, "__metadata:") || strings.Contains(content, "version:") {
		// Try to parse as YAML to confirm Berry format
		var yamlContent interface{}
		if err := yaml.Unmarshal(buffer[:n], &yamlContent); err == nil {
			return true, nil
		}
	}

	// If we can't determine from content, default to Classic
	return false, nil
}

// GetYarnrcConfig reads and returns the .yarnrc.yml configuration if it exists
func (d *YarnDetector) GetYarnrcConfig() (map[string]interface{}, error) {
	yarnrcPath := filepath.Join(d.projectPath, ".yarnrc.yml")
	
	exists, err := fs.Exists(yarnrcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check for .yarnrc.yml: %w", err)
	}
	
	if !exists {
		return make(map[string]interface{}), nil
	}

	content, err := os.ReadFile(yarnrcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .yarnrc.yml: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse .yarnrc.yml: %w", err)
	}

	return config, nil
}
