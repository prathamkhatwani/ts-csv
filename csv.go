package csv

import (
	"fmt"
)

// Library metadata
const (
	LibName    = "@gregoranders/csv"
	LibVersion = "0.0.13"
	LibURL     = "https://gregoranders.github.io/ts-csv/"
)

// Field represents a CSV field value
type Field = string

// Row represents a CSV row containing multiple fields
type Row = []Field

// Configuration for parser delimiters and quote character
type Configuration struct {
	FieldSeparator string
	LineSeparator  string
	Quote          string
}

// DefaultConfiguration defines the default parser options
var DefaultConfiguration = Configuration{
	FieldSeparator: ",",
	LineSeparator:  "\n",
	Quote:          "\"",
}

// State tracks the state of the CSV parser
type State struct {
	AppendCell  bool
	AppendField bool
	AppendRow   bool
	Field       int
	FieldOffset int
	Line        int
	LineOffset  int
	Quoted      bool
}

var csvInitialState = State{
	Field:        0,
	FieldOffset:  0,
	Line:         0,
	LineOffset:   0,
	Quoted:       false,
	AppendCell:   false,
	AppendField:  false,
	AppendRow:    false,
}

// ParseError is returned when the CSV structure is malformed
type ParseError struct {
	Line   int
	Column int
}

// Error implements the error interface for ParseError
func (e *ParseError) Error() string {
	return fmt.Sprintf("Invalid CSV at %d:%d", e.Line, e.Column)
}

// Parser manages the state and logic for CSV parsing
type Parser struct {
	rows       []Row
	row        Row
	cell       string
	options    Configuration
	state      State
	index      int
	current    string
	previous   string
	quoteState State
}

// NewParser creates a new Parser with optional configuration overrides
func NewParser(config ...Configuration) *Parser {
	opts := DefaultConfiguration
	if len(config) > 0 {
		cfg := config[0]
		if cfg.FieldSeparator != "" {
			opts.FieldSeparator = cfg.FieldSeparator
		}
		if cfg.LineSeparator != "" {
			opts.LineSeparator = cfg.LineSeparator
		}
		if cfg.Quote != "" {
			opts.Quote = cfg.Quote
		}
	}
	return &Parser{
		options:    opts,
		rows:       []Row{},
		row:        Row{},
		state:      csvInitialState,
		quoteState: csvInitialState,
	}
}

// Parse parses the given CSV text and returns rows or an error
func (p *Parser) Parse(text string) ([]Row, error) {
	p.Reset()

	runes := []rune(text)
	for p.index = 0; p.index < len(runes); p.index++ {
		p.state.AppendCell = true
		p.previous = p.current
		p.current = string(runes[p.index])
		if err := p.handleNext(); err != nil {
			return nil, err
		}
	}

	if len(p.row) > 0 {
		p.addField(p.fieldValue(p.cell))
		p.addRow()
	}

	if p.state.Quoted {
		return nil, &ParseError{
			Line:   p.quoteState.Line,
			Column: p.quoteState.LineOffset,
		}
	}

	return p.rows, nil
}

// Rows returns the parsed rows
func (p *Parser) Rows() []Row {
	return p.rows
}

// JSON maps the parsed rows to maps of key-value pairs using the first row as headers
func (p *Parser) JSON() []map[string]string {
	if len(p.rows) == 0 {
		return []map[string]string{}
	}

	keys := p.rows[0]
	result := []map[string]string{}

	for index := 1; index < len(p.rows); index++ {
		row := p.rows[index]
		item := make(map[string]string)
		for keyIndex, key := range keys {
			if keyIndex < len(row) {
				item[key] = row[keyIndex]
			} else {
				item[key] = ""
			}
		}
		result = append(result, item)
	}

	return result
}

func (p *Parser) handleNext() error {
	quoted, err := p.handleQuote()
	if err != nil {
		return err
	}
	if !quoted {
		if !p.handleFieldSeparator() {
			p.handleLineSeparator()
		}
	}
	p.processState()
	return nil
}

func (p *Parser) handleQuote() (bool, error) {
	if p.current == p.options.Quote {
		p.quoteState = p.state
		if p.previous == "\\" {
			p.handleQuoteEscaped()
		} else {
			if err := p.handleQuoteNotEscaped(); err != nil {
				return true, err
			}
		}
		return true, nil
	}
	return false, nil
}

func (p *Parser) handleQuoteEscaped() {
	if len(p.cell) > 0 {
		p.cell = p.cell[:len(p.cell)-1]
	}
}

func (p *Parser) handleQuoteNotEscaped() error {
	if len(p.cell) == 0 || p.state.Quoted {
		p.state.Quoted = !p.state.Quoted
	} else {
		return &ParseError{
			Line:   p.state.Line,
			Column: p.state.LineOffset,
		}
	}
	p.state.AppendCell = false
	return nil
}

func (p *Parser) handleFieldSeparator() bool {
	if p.current == p.options.FieldSeparator {
		if !p.state.Quoted {
			p.state.AppendCell = false
			p.state.AppendField = true
		}
		return true
	}
	return false
}

func (p *Parser) handleLineSeparator() bool {
	if p.current == p.options.LineSeparator {
		if !p.state.Quoted {
			p.state.AppendCell = false
			p.state.AppendField = true
			p.state.AppendRow = true
		}
		return true
	}
	return false
}

func (p *Parser) processState() {
	if p.state.AppendCell {
		p.cell += p.current
	}

	if p.state.AppendField {
		p.addField(p.fieldValue(p.cell))
		p.cell = ""
	}

	if p.state.AppendRow {
		p.addRow()
		p.row = Row{}
	}

	p.state.LineOffset++
	p.state.FieldOffset++
}

func (p *Parser) fieldValue(cell string) string {
	return cell
}

func (p *Parser) addField(field string) {
	p.row = append(p.row, field)
	p.state.Field++
	p.state.FieldOffset = -1
	p.state.AppendField = false
}

func (p *Parser) addRow() {
	p.rows = append(p.rows, p.row)
	p.state.Field = 0
	p.state.Line++
	p.state.LineOffset = -1
	p.state.AppendRow = false
}

// Reset resets the parser back to its initial state
func (p *Parser) Reset() {
	p.rows = []Row{}
	p.row = Row{}
	p.cell = ""
	p.state = csvInitialState
	p.index = 0
	p.current = ""
	p.previous = ""
	p.quoteState = csvInitialState
}
