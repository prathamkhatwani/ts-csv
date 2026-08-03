// Bundled version of original ts-csv parser for differential fuzzing
const DefaultConfiguration = {
  fieldSeparator: ',',
  lineSeparator: '\n',
  quote: '"',
};

const CSV_INITIAL_STATE = {
  field: 0,
  fieldOffset: 0,
  line: 0,
  lineOffset: 0,
  quoted: false,
  appendCell: false,
  appendField: false,
  appendRow: false,
};

class ParseError extends Error {
  constructor(line, column) {
    super(`Invalid CSV at ${line}:${column}`);
  }
}

class Parser {
  constructor(configuration = DefaultConfiguration) {
    this._rows = [];
    this._row = [];
    this._cell = '';
    this._options = Object.assign({}, DefaultConfiguration, configuration);
    this._state = { ...CSV_INITIAL_STATE };
    this._index = 0;
    this._current = '';
    this._previous = '';
    this._quoteState = { ...CSV_INITIAL_STATE };
  }

  parse(text) {
    this.reset();

    for (this._index = 0; this._index < text.length; this._index++) {
      this._state.appendCell = true;
      this._previous = this._current;
      this._current = text[this._index];
      this.handleNext();
    }

    if (this._row.length > 0) {
      this.addField(this.fieldValue(this._cell), this._row, this._state);
      this.addRow(this._row, this._rows, this._state);
    }

    if (this._state.quoted) {
      throw new ParseError(
        this._quoteState.line,
        this._quoteState.lineOffset,
      );
    }

    this.makeImmutable();

    return this._rows;
  }

  get rows() {
    return this._rows;
  }

  get json() {
    if (this.rows.length > 0) {
      const keys = this.rows[0].filter(
        (field) => typeof field === 'string',
      );

      return Object.freeze(
        this.rows
          .filter((row, index) => row && index > 0)
          .map((row) => {
            const object = {};
            keys.forEach((key, keyIndex) => {
              object[key] = row[keyIndex];
            });
            return Object.freeze(object);
          }),
      );
    }

    return Object.freeze([]);
  }

  handleNext() {
    this.handleQuote() ||
      this.handleFieldSeparator() ||
      this.handleLineSeparator();
    this.processState();
  }

  handleQuote() {
    if (this._current === this._options.quote) {
      this._quoteState = { ...this._state };
      if (this._previous === '\\') {
        this.handleQuoteEscaped();
      } else {
        this.handleQuoteNotEscaped();
      }
      return true;
    }
    return false;
  }

  handleQuoteEscaped() {
    this._cell = this._cell.slice(0, Math.max(0, this._cell.length - 1));
  }

  handleQuoteNotEscaped() {
    if (this._cell.length === 0 || this._state.quoted) {
      this._state.quoted = !this._state.quoted;
    } else {
      throw new ParseError(this._state.line, this._state.lineOffset);
    }
    this._state.appendCell = false;
  }

  handleFieldSeparator() {
    if (this._current === this._options.fieldSeparator) {
      if (!this._state.quoted) {
        this._state.appendCell = false;
        this._state.appendField = true;
      }
      return true;
    }
    return false;
  }

  handleLineSeparator() {
    if (this._current === this._options.lineSeparator) {
      if (!this._state.quoted) {
        this._state.appendCell = false;
        this._state.appendField = true;
        this._state.appendRow = true;
      }
      return true;
    }
    return false;
  }

  processState() {
    if (this._state.appendCell) {
      this._cell += this._current;
    }

    if (this._state.appendField) {
      this.addField(this.fieldValue(this._cell), this._row, this._state);
      this._cell = '';
    }

    if (this._state.appendRow) {
      this.addRow(this._row, this._rows, this._state);
      this._row = [];
    }

    this._state.lineOffset++;
    this._state.fieldOffset++;
  }

  fieldValue(cell) {
    return cell;
  }

  addField(field, row, state) {
    row.push(field);
    state.field++;
    state.fieldOffset = -1;
    state.appendField = false;
  }

  addRow(row, rows, state) {
    rows.push(row);
    state.field = 0;
    state.line++;
    state.lineOffset = -1;
    state.appendRow = false;
  }

  makeImmutable() {
    this.rows.forEach((row) => {
      row.forEach((value) => Object.freeze(value));
      Object.freeze(row);
    });
    Object.freeze(this._rows);
  }

  reset() {
    this._rows = [];
    this._row = [];
    this._cell = '';
    this._state = { ...CSV_INITIAL_STATE };
    this._index = 0;
    this._current = '';
    this._previous = '';
    this._quoteState = { ...CSV_INITIAL_STATE };
  }
}

module.exports = { Parser };
