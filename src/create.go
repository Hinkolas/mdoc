package src

import "html/template"

type Document struct {
	input    *Input
	template *template.Template
}

type Input struct {
	config map[string]any
	pages  []string
}

// TODO: Generates the pdf and saves it to the given path.
func (d *Document) Save(path string) error {
	return nil
}

func Create(input string) (*Document, error) {

	t, err := template.ParseFiles("page.html")
	if err != nil {
		return nil, err
	}

	i, err := parseInput(input)
	if err != nil {
		return nil, err
	}

	return &Document{i, t}, nil

}

// Creates
func CreateWithTemplate(input, template string) {}

// Creates an text object from the given input string.
// Enables special syntax like YAML front matter and manual page-break.
func parseInput(input string) (*Input, error) {

	// TODO:
	// This function should take the raw markdown input and parse it into a easy
	// to render format. It should split the input at manual page-breaks and
	// cut the config yaml front matter from the beginning of the markdown.
	return &Input{
		config: make(map[string]any),
		pages:  []string{},
	}, nil

}
