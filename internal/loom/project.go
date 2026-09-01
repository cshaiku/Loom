package loom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ProjectBuildReport struct {
	SchemaVersion string                 `json:"schema_version"`
	Status        string                 `json:"status"`
	Project       string                 `json:"project"`
	ManifestPath  string                 `json:"manifestPath"`
	ProjectRoot   string                 `json:"projectRoot"`
	OutputDir     string                 `json:"outputDir"`
	Artifacts     []ProjectBuildArtifact `json:"artifacts"`
	Diagnostics   []Diagnostic           `json:"diagnostics"`
}

type ProjectBuildArtifact struct {
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Status string `json:"status"`
}

func ProjectBuild(manifestPath, projectRoot, outputDir string, overwrite bool) (ProjectBuildReport, error) {
	manifestPath = filepath.Clean(manifestPath)
	if projectRoot == "" {
		projectRoot = filepath.Dir(manifestPath)
	}
	projectRoot = filepath.Clean(projectRoot)
	if outputDir == "" {
		outputDir = filepath.Join(projectRoot, "generated", "loom-build")
	}
	outputDir = filepath.Clean(outputDir)

	report := ProjectBuildReport{
		SchemaVersion: "1",
		Status:        "ok",
		ManifestPath:  manifestPath,
		ProjectRoot:   projectRoot,
		OutputDir:     outputDir,
		Artifacts:     []ProjectBuildArtifact{},
		Diagnostics:   []Diagnostic{},
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return report, err
	}

	validation := DiagnosticsProjectConfigValidate(manifestPath, projectRoot)
	report.Project = validation.Project
	if err := writeProjectJSONArtifact(&report, outputDir, "manifest-validation.json", "manifest-validation", validation.Status, validation, overwrite); err != nil {
		return report, err
	}
	if validation.Status != "ok" {
		report.Status = "error"
		report.Diagnostics = append(report.Diagnostics, manifestValidationDiagnostics(validation)...)
		_ = writeProjectJSONArtifact(&report, outputDir, "project-summary.json", "project-summary", report.Status, report, overwrite)
		return report, ErrCommandFailed
	}

	manifest, err := readLoomManifest(manifestPath)
	if err != nil {
		return report, err
	}
	report.Project = manifest.Project
	sourcePath := resolveManifestPath(manifest.Source, projectRoot)
	sourcePlatform := InferSourcePlatform(sourcePath)
	targetPlatform := firstNonEmpty(manifest.Target, defaultTransferTarget(sourcePlatform))

	sourceAnalysis, err := AnalyzeByPlatform(sourcePath, sourcePlatform)
	if err != nil {
		report.Status = "error"
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityError, Code: "PROJECT.SOURCE_ANALYSIS", Message: err.Error()})
		_ = writeProjectJSONArtifact(&report, outputDir, "project-summary.json", "project-summary", report.Status, report, overwrite)
		return report, err
	}
	if err := writeProjectJSONArtifact(&report, outputDir, "source-analysis.json", "source-analysis", statusFromDiagnostics(sourceAnalysis.Diagnostics), sourceAnalysis, overwrite); err != nil {
		return report, err
	}
	contracts := GenerateContracts(sourceAnalysis, targetPlatform)
	if err := writeProjectJSONArtifact(&report, outputDir, "target-contracts.json", "target-contracts", contracts.Status, contracts, overwrite); err != nil {
		return report, err
	}
	if targetPlatform == "winui3" || targetPlatform == "windows" {
		xaml := GenerateXAML(sourceAnalysis, manifest.ThemeResourcePrefix, false)
		if err := writeProjectTextArtifact(&report, outputDir, "generated.xaml", "generated-xaml", xaml.Status, xaml.Text, overwrite); err != nil {
			return report, err
		}
		if err := writeProjectJSONArtifact(&report, outputDir, "generated-xaml-report.json", "generated-xaml-report", xaml.Status, xaml, overwrite); err != nil {
			return report, err
		}
	}
	swift := GenerateSwiftUI(sourceAnalysis, manifest.RootView)
	if err := writeProjectTextArtifact(&report, outputDir, "GeneratedView.swift", "generated-swiftui", swift.Status, swift.Text, overwrite); err != nil {
		return report, err
	}
	if err := writeProjectJSONArtifact(&report, outputDir, "generated-swiftui-report.json", "generated-swiftui-report", swift.Status, swift, overwrite); err != nil {
		return report, err
	}

	graph, err := GraphComponents(projectRoot, manifest.RootView, "", "", "")
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityWarning, Code: "PROJECT.COMPONENT_GRAPH", Message: err.Error()})
	} else if err := writeProjectJSONArtifact(&report, outputDir, "component-graph.json", "component-graph", graph.Status, graph, overwrite); err != nil {
		return report, err
	}

	patterns, err := LoadPatterns(DefaultPatternDirectory)
	if err != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityWarning, Code: "PROJECT.PATTERNS", Message: err.Error()})
	} else {
		transfer := Transfer(sourceAnalysis, patterns, sourcePlatform, targetPlatform)
		if err := writeProjectJSONArtifact(&report, outputDir, "source-transfer.json", "source-transfer", transfer.Status, transfer, overwrite); err != nil {
			return report, err
		}
	}

	if manifest.ExistingXaml != "" {
		xamlPath := resolveManifestPath(manifest.ExistingXaml, projectRoot)
		xamlAnalysis, err := AnalyzeByPlatform(xamlPath, "winui3")
		if err != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityWarning, Code: "PROJECT.EXISTING_XAML", Message: err.Error()})
		} else if err := writeProjectJSONArtifact(&report, outputDir, "existing-xaml-analysis.json", "existing-xaml-analysis", statusFromDiagnostics(xamlAnalysis.Diagnostics), xamlAnalysis, overwrite); err != nil {
			return report, err
		}
		parity, err := InspectParity(sourcePath, xamlPath, sourcePlatform, "winui3")
		if err != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityWarning, Code: "PROJECT.SOURCE_EXISTING_PARITY", Message: err.Error()})
		} else if err := writeProjectJSONArtifact(&report, outputDir, "source-existing-parity.json", "source-existing-parity", parity.Status, parity, overwrite); err != nil {
			return report, err
		}
	}

	if manifest.ReferenceLayout != "" {
		referencePath := resolveManifestPath(manifest.ReferenceLayout, projectRoot)
		referencePlatform := InferSourcePlatform(referencePath)
		referenceAnalysis, err := AnalyzeByPlatform(referencePath, referencePlatform)
		if err != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityWarning, Code: "PROJECT.REFERENCE_ANALYSIS", Message: err.Error()})
		} else if err := writeProjectJSONArtifact(&report, outputDir, "reference-analysis.json", "reference-analysis", statusFromDiagnostics(referenceAnalysis.Diagnostics), referenceAnalysis, overwrite); err != nil {
			return report, err
		}
		parity, err := InspectParity(sourcePath, referencePath, sourcePlatform, referencePlatform)
		if err != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityWarning, Code: "PROJECT.SOURCE_REFERENCE_PARITY", Message: err.Error()})
		} else if err := writeProjectJSONArtifact(&report, outputDir, "source-reference-parity.json", "source-reference-parity", parity.Status, parity, overwrite); err != nil {
			return report, err
		}
	}

	report.Status = projectBuildStatus(report)
	if err := writeProjectJSONArtifact(&report, outputDir, "project-summary.json", "project-summary", report.Status, report, overwrite); err != nil {
		return report, err
	}
	if report.Status == "error" {
		return report, ErrCommandFailed
	}
	return report, nil
}

func readLoomManifest(path string) (LoomManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LoomManifest{}, err
	}
	manifest := LoomManifest{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return LoomManifest{}, err
	}
	return manifest, nil
}

func writeProjectJSONArtifact(report *ProjectBuildReport, outputDir, name, kind, status string, value any, overwrite bool) error {
	text, err := prettyJSON(value)
	if err != nil {
		return err
	}
	return writeProjectTextArtifact(report, outputDir, name, kind, status, text, overwrite)
}

func writeProjectTextArtifact(report *ProjectBuildReport, outputDir, name, kind, status string, text string, overwrite bool) error {
	path := filepath.Join(outputDir, name)
	artifact := ProjectBuildArtifact{Kind: kind, Path: path, Status: status}
	report.Artifacts = append(report.Artifacts, artifact)
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
			return fmt.Errorf("refusing to overwrite existing output %s without --overwrite", path)
		} else if !os.IsNotExist(err) {
			report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
			return err
		}
	}
	temp, err := os.CreateTemp(outputDir, "."+name+".tmp-*")
	if err != nil {
		report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.WriteString(text); err != nil {
		temp.Close()
		report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
		return err
	}
	if err := temp.Close(); err != nil {
		report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
		return err
	}
	if err := os.Rename(tempName, path); err != nil {
		report.Artifacts = report.Artifacts[:len(report.Artifacts)-1]
		return err
	}
	return nil
}

func manifestValidationDiagnostics(report LoomManifestValidationReport) []Diagnostic {
	diagnostics := []Diagnostic{}
	for _, issue := range report.Issues {
		diagnostics = append(diagnostics, Diagnostic{Severity: issue.Severity, Code: issue.Code, Message: issue.Detail})
	}
	return diagnostics
}

func statusFromDiagnostics(diagnostics []Diagnostic) string {
	status := "ok"
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			return "error"
		}
		if diagnostic.Severity == SeverityWarning {
			status = "warning"
		}
	}
	return status
}

func projectBuildStatus(report ProjectBuildReport) string {
	status := "ok"
	for _, artifact := range report.Artifacts {
		if artifact.Status == "error" || artifact.Status == "source-invalid" {
			return "error"
		}
		if artifact.Status != "ok" {
			status = "warning"
		}
	}
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Severity == SeverityError {
			return "error"
		}
		if diagnostic.Severity == SeverityWarning {
			status = "warning"
		}
	}
	return status
}

func ProjectBuildText(report ProjectBuildReport) string {
	return fmt.Sprintf("loom project build\nstatus: %s\nproject: %s\noutput: %s\nartifacts: %d\n", report.Status, report.Project, report.OutputDir, len(report.Artifacts))
}
