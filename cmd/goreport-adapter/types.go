package main

type TestCase struct {
	Package string
	Name    string
	Status  string
	Elapsed float64
	Output  string
}

type CoverageLine struct {
	Path string
	Line int
	Hits int
}
