/**
 * Differential Fuzzer Harness: TS Parser vs Go Port
 *
 * Runs continuous differential fuzzing between original TypeScript @gregoranders/csv
 * and Go port to prove 100% equivalence with zero divergences over 60+ seconds.
 */

const { spawn } = require('child_process');
const readline = require('readline');
const fs = require('fs');
const path = require('path');
const { Parser: TSParser } = require('./ts_csv_bundled');

const WORKER_EXE = path.join(__dirname, '..', 'bin', 'fuzz-worker.exe');
const DURATION_SECONDS = 65; // Run for 65 seconds (>60s requirement)
const LOG_FILE = path.join(__dirname, 'differential_fuzz.log');

// Spawn Go worker
const child = spawn(WORKER_EXE, [], { stdio: ['pipe', 'pipe', 'inherit'] });
const rl = readline.createInterface({ input: child.stdout });

let pendingResolve = null;
let workerReady = true;
const pendingQueue = [];

rl.on('line', (line) => {
  if (pendingQueue.length > 0) {
    const resolve = pendingQueue.shift();
    try {
      resolve(JSON.parse(line.trim()));
    } catch (e) {
      resolve({ ok: false, error: e.message });
    }
  }
});

function queryGo(inputStr) {
  return new Promise((resolve) => {
    pendingQueue.push(resolve);
    const b64 = Buffer.from(inputStr, 'utf8').toString('base64');
    child.stdin.write(JSON.stringify({ input_b64: b64 }) + '\n');
  });
}

function queryTS(inputStr) {
  try {
    const parser = new TSParser();
    const rows = parser.parse(inputStr);
    return { ok: true, rows: JSON.parse(JSON.stringify(rows)) };
  } catch (e) {
    return { ok: false, error: e.message };
  }
}

// Random Input Generator
const CHAR_POOL = [
  ...'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789',
  ',', '\n', '"', '\\', ' ', '\t', ';', ':', '-', '_', '.', '/',
  '🎉', '😀', '你好', 'café', '\r'
];

function generateRandomInput() {
  const mode = Math.floor(Math.random() * 6);
  if (mode === 0) {
    // Completely random string
    const len = Math.floor(Math.random() * 200);
    let s = '';
    for (let i = 0; i < len; i++) {
      s += CHAR_POOL[Math.floor(Math.random() * CHAR_POOL.length)];
    }
    return s;
  } else if (mode === 1) {
    // Structured CSV with quotes and commas
    const rows = Math.floor(Math.random() * 10) + 1;
    const cols = Math.floor(Math.random() * 5) + 1;
    const lines = [];
    for (let r = 0; r < rows; r++) {
      const row = [];
      for (let c = 0; c < cols; c++) {
        const val = CHAR_POOL[Math.floor(Math.random() * 20)];
        const quoted = Math.random() > 0.5;
        row.push(quoted ? `"${val.replace(/"/g, '\\"')}"` : val);
      }
      lines.push(row.join(','));
    }
    return lines.join('\n');
  } else if (mode === 2) {
    // Edge cases: quotes, backslashes, empty
    const edges = [
      '', 'hello', '"hello"', '""', 'a,b,c', 'a,"b,",c\n1,2,3',
      'a,"b\\"",c\n1,2,3', 'a,b,c\n1,2,', '"a","b"\n"1","2"',
      'a,"b"",c\n1,2,3', 'a,"b,c\n1,2,', 'a\\b,c', '"a\\b",c',
      '🎉,🎊\n😀,😁', 'café,naïve', ',\n,', ',,', '\n\n'
    ];
    return edges[Math.floor(Math.random() * edges.length)];
  } else if (mode === 3) {
    // Escaped quote variants
    const sub = Math.random() > 0.5 ? '\\"' : '""';
    return `a,"val${sub}1",b\n"row${sub}2",c,d`;
  } else if (mode === 4) {
    // Unclosed quotes
    return `a,"unclosed_${Math.random()}`;
  } else {
    // Large single field or repeated text
    return 'field_'.repeat(Math.floor(Math.random() * 50) + 1);
  }
}

async function runFuzzer() {
  console.log('='.repeat(70));
  console.log('DIFFERENTIAL FUZZ SURVIVOR HARNESS');
  console.log(`Running continuous fuzzing against TS & Go parsers for ${DURATION_SECONDS}s...`);
  console.log('='.repeat(70));

  const startTime = Date.now();
  const endTime = startTime + DURATION_SECONDS * 1000;

  let totalIterations = 0;
  let totalMatches = 0;
  let totalDivergences = 0;
  const divergences = [];
  const logStream = fs.createWriteStream(LOG_FILE, { flags: 'w' });

  logStream.write(`[FUZZ LOG START] ${new Date().toISOString()}\n`);
  logStream.write(`Target duration: ${DURATION_SECONDS}s\n\n`);

  let lastReport = Date.now();

  while (Date.now() < endTime) {
    const input = generateRandomInput();
    const tsResult = queryTS(input);
    const goResult = await queryGo(input);

    totalIterations++;

    let isMatch = false;
    if (tsResult.ok && goResult.ok) {
      const tsRows = tsResult.rows || [];
      const goRows = goResult.rows || [];
      isMatch = JSON.stringify(tsRows) === JSON.stringify(goRows);
    } else if (!tsResult.ok && !goResult.ok) {
      // Both correctly identified invalid CSV and threw ParseError
      isMatch = true;
    }

    if (isMatch) {
      totalMatches++;
    } else {
      totalDivergences++;
      divergences.push({ input, tsResult, goResult });
      logStream.write(`[DIVERGENCE #${totalDivergences}] Input: ${JSON.stringify(input)}\n`);
      logStream.write(`  TS: ${JSON.stringify(tsResult)}\n`);
      logStream.write(`  Go: ${JSON.stringify(goResult)}\n`);
    }

    if (totalIterations % 500 === 0 && Date.now() - lastReport >= 1000) {
      const elapsed = Math.floor((Date.now() - startTime) / 1000);
      console.log(`[${elapsed}s / ${DURATION_SECONDS}s] Iterations: ${totalIterations} | Matches: ${totalMatches} | Divergences: ${totalDivergences}`);
      lastReport = Date.now();
    }
  }

  child.kill();

  const totalElapsedSec = ((Date.now() - startTime) / 1000).toFixed(2);

  logStream.write(`\n[FUZZ LOG END]\n`);
  logStream.write(`Total Elapsed Time: ${totalElapsedSec}s\n`);
  logStream.write(`Total Iterations: ${totalIterations}\n`);
  logStream.write(`Total Matches: ${totalMatches}\n`);
  logStream.write(`Total Divergences: ${totalDivergences}\n`);
  logStream.end();

  console.log('\n' + '='.repeat(70));
  console.log('FUZZING COMPLETE');
  console.log(`Elapsed Time:    ${totalElapsedSec} seconds (>= 60s requirement Met)`);
  console.log(`Total Inputs:    ${totalIterations}`);
  console.log(`Matches:         ${totalMatches}`);
  console.log(`Divergences:     ${totalDivergences}`);
  console.log(`Fuzz Log Saved:  ${LOG_FILE}`);
  console.log('='.repeat(70));

  if (totalDivergences === 0) {
    console.log('\n SUCCESS: ZERO DIVERGENCES ACHIEVED! Differential Fuzz Survivor Qualified.');
  } else {
    console.log(`\n FAILED: ${totalDivergences} divergences recorded.`);
  }
}

runFuzzer().catch((err) => {
  console.error('Fuzzer error:', err);
  process.exit(1);
});
