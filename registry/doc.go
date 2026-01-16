// Package registry provides a widget registration system for dynamic widget creation.
//
// The registry allows third-party developers to register custom widgets that can
// be instantiated by name at runtime. This enables dynamic widget creation from
// configuration files, declarative UI builders, or visual editors.
//
// # Quick Start
//
// Register a widget factory during package initialization:
//
//	package mywidgets
//
//	import "github.com/gogpu/ui/registry"
//
//	func init() {
//	    registry.RegisterWidget("my-button", NewMyButton, registry.WidgetInfo{
//	        Name:        "my-button",
//	        Description: "Custom button widget",
//	        Category:    registry.CategoryInput,
//	        Version:     "1.0.0",
//	    })
//	}
//
//	func NewMyButton(config map[string]any) (registry.Widget, error) {
//	    label, _ := config["label"].(string)
//	    return &MyButton{label: label}, nil
//	}
//
// Create widgets dynamically in your application:
//
//	package main
//
//	import (
//	    "github.com/gogpu/ui/registry"
//	    _ "github.com/example/mywidgets" // Auto-registers via init()
//	)
//
//	func main() {
//	    // Create widget by name
//	    widget, err := registry.CreateWidget("my-button", map[string]any{
//	        "label": "Click Me",
//	    })
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // List all registered widgets
//	    for _, name := range registry.ListWidgets() {
//	        info, _ := registry.GetWidgetInfo(name)
//	        fmt.Printf("%s (%s): %s\n", info.Name, info.Category, info.Description)
//	    }
//	}
//
// # Thread Safety
//
// The registry is safe for concurrent access. Multiple goroutines can register,
// create, and query widgets simultaneously without external synchronization.
//
// # Categories
//
// Widgets are organized into categories for discoverability:
//
//   - CategoryInput: Interactive widgets (buttons, text fields, checkboxes)
//   - CategoryDisplay: Read-only widgets (labels, images, progress bars)
//   - CategoryContainer: Widgets that contain other widgets (panels, scroll views)
//   - CategoryCustom: User-defined widgets that don't fit other categories
//
// # Best Practices
//
// Widget factories should:
//
//  1. Validate all required configuration parameters
//  2. Provide sensible defaults for optional parameters
//  3. Return descriptive errors for invalid configuration
//  4. Initialize widgets to a valid state
//
// Example with validation:
//
//	func NewValidatedWidget(config map[string]any) (registry.Widget, error) {
//	    // Validate required parameters
//	    label, ok := config["label"].(string)
//	    if !ok || label == "" {
//	        return nil, fmt.Errorf("label is required and must be a non-empty string")
//	    }
//
//	    // Optional parameter with default
//	    width, _ := config["width"].(float64)
//	    if width <= 0 {
//	        width = 100 // Default width
//	    }
//
//	    return &MyWidget{label: label, width: width}, nil
//	}
package registry
