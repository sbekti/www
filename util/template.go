package util

import (
	"html/template"
	"path/filepath"
)

// LoadTemplatesFromDirectory loads all HTML templates from a directory and its subdirectories,
// preserving the directory structure in template names
func LoadTemplatesFromDirectory(baseDir string) (*template.Template, error) {
	templates := template.New("")
	
	// Get all HTML files in the base directory
	mainTemplates, err := filepath.Glob(filepath.Join(baseDir, "*.html"))
	if err != nil {
		return nil, err
	}
	
	// Get all HTML files in subdirectories (one level deep)
	subDirPattern := filepath.Join(baseDir, "*", "*.html")
	subTemplates, err := filepath.Glob(subDirPattern)
	if err != nil {
		return nil, err
	}
	
	// Combine all template files
	allTemplateFiles := append(mainTemplates, subTemplates...)
	
	for _, file := range allTemplateFiles {
		// Get relative path from base directory
		relPath, err := filepath.Rel(baseDir, file)
		if err != nil {
			return nil, err
		}
		
		// Normalize path separators for template names
		templateName := filepath.ToSlash(relPath)
		
		// Read and parse the template file
		tmplContent, err := template.ParseFiles(file)
		if err != nil {
			return nil, err
		}
		
		// Get the parsed template (it will have the filename as name)
		parsedTemplate := tmplContent.Lookup(filepath.Base(file))
		if parsedTemplate == nil {
			continue // Skip if template not found
		}
		
		// Add the template with the relative path as name
		_, err = templates.AddParseTree(templateName, parsedTemplate.Tree)
		if err != nil {
			return nil, err
		}
	}
	
	return templates, nil
} 