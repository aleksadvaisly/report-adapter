package main

import (
	"encoding/xml"
	"fmt"
)

type coberturaCoverage struct {
	Packages []coberturaPackage `xml:"packages>package"`
	Classes  []coberturaClass   `xml:"classes>class"`
}

type coberturaPackage struct {
	Classes []coberturaClass `xml:"classes>class"`
}

type coberturaClass struct {
	FileName string          `xml:"filename,attr"`
	Lines    []coberturaLine `xml:"lines>line"`
}

type coberturaLine struct {
	Number int `xml:"number,attr"`
	Hits   int `xml:"hits,attr"`
}

func parseCobertura(data []byte) ([]CoverageLine, error) {
	var report coberturaCoverage
	if err := xml.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse cobertura xml: %w", err)
	}

	classes := append([]coberturaClass{}, report.Classes...)
	for _, pkg := range report.Packages {
		classes = append(classes, pkg.Classes...)
	}

	lines := make([]CoverageLine, 0)
	for _, class := range classes {
		for _, line := range class.Lines {
			lines = append(lines, CoverageLine{
				Path: class.FileName,
				Line: line.Number,
				Hits: line.Hits,
			})
		}
	}
	return lines, nil
}
