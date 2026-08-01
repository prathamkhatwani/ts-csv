/* eslint-disable */
/**
 * simple CSV parser (Go Thin Adapter)
 *
 * @packageDocumentation
 */

import { execFileSync } from 'child_process';
import * as path from 'path';

/**
 * library name
 *
 * @public
 * @readonly
 */
export const libname = '@gregoranders/csv';

/**
 * library version
 *
 * @public
 * @readonly
 */
export const libversion = '0.0.13';

/**
 * library homepage
 *
 * @public
 * @readonly
 */
export const liburl = 'https://gregoranders.github.io/ts-csv/';

/**
 * csv field
 *
 * @public
 */
export type Field = string;

/**
 * csv row
 *
 * @public
 */
export type Row = Array<Field>;

/**
 * parser configuration
 *
 * @public
 */
export interface Configuration {
  fieldSeparator?: string;
  lineSeparator?: string;
  quote?: string;
}

const DefaultConfiguration = {
  fieldSeparator: ',',
  lineSeparator: '\n',
  quote: '"',
};

class ParseError extends Error {
  public constructor(message: string) {
    super(message);
  }
}

/**
 * csv parser
 *
 * @public
 */
export class Parser<T = Record<string, string>> {
  private _rows: Row[] = [];
  private _json: T[] = [];
  private _options: Configuration;

  public constructor(configuration: Configuration = DefaultConfiguration) {
    this._options = Object.assign({}, DefaultConfiguration, configuration);
  }

  public parse(text: string): readonly Row[] {
    const req = {
      fieldSeparator: this._options.fieldSeparator,
      lineSeparator: this._options.lineSeparator,
      quote: this._options.quote,
      text: text,
    };

    const binPath = path.resolve(__dirname, '..', '..', 'bin', 'csv-cli');
    
    // Pass the configuration and input text to the Go binary via stdin
    const input = JSON.stringify(req);
    
    let stdout: Buffer;
    try {
      stdout = execFileSync(binPath, [], { input: input, maxBuffer: 10 * 1024 * 1024 });
    } catch (err: any) {
      throw new Error(`Failed to execute Go binary: ${err.message}`);
    }

    const res = JSON.parse(stdout.toString());

    if (res.error) {
      throw new ParseError(res.error);
    }

    this._rows = res.rows || [];
    this._json = res.json || [];
    
    this.makeImmutable();
    return this._rows;
  }

  public get rows(): readonly Row[] {
    return this._rows;
  }

  public get json(): readonly T[] {
    return this._json;
  }

  private makeImmutable() {
    this._rows.forEach((row) => {
      Object.freeze(row);
    });
    Object.freeze(this._rows);

    this._json.forEach((obj) => {
      Object.freeze(obj);
    });
    Object.freeze(this._json);
  }
}

export default Parser;
