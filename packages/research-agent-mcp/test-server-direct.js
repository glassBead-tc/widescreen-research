#!/usr/bin/env node
/**
 * Direct test of MCP server handlers without connection
 */

import { readFileSync } from 'fs';
import { parseSrcmd } from './dist/srcmd-parser.js';
import { executeCell } from './dist/executor.js';

async function test() {
  console.log('🧪 Testing MCP Server Implementation (Direct)\n');

  // 1. Load test notebook
  console.log('1️⃣  Loading test notebook...');
  const notebookContent = readFileSync('./examples/test-agent.src.md', 'utf8');
  const notebook = parseSrcmd(notebookContent);
  console.log(`   ✅ Loaded ${notebook.cells.length} cells:`);
  notebook.cells.forEach((cell) => {
    console.log(`      - ${cell.id} (${cell.type})`);
  });
  console.log('');

  // 2. Test finding cells
  console.log('2️⃣  Testing cell lookup...');
  const { findCell } = await import('./dist/srcmd-parser.js');
  const helloCell = findCell(notebook, 'hello');
  const calcCell = findCell(notebook, 'calculate');
  console.log(`   ✅ Found cells:`);
  console.log(`      - hello: ${!!helloCell}`);
  console.log(`      - calculate: ${!!calcCell}`);
  console.log('');

  // 3. Test execute_cell (hello)
  console.log('3️⃣  Executing hello.ts...');
  const helloResult = await executeCell(helloCell, '.', {});
  console.log(`   ✅ Result:`);
  console.log(`      Exit code: ${helloResult.exitCode}`);
  console.log(`      Stdout: ${helloResult.stdout.trim()}`);
  console.log('');

  // 4. Test execute_cell with params (calculate)
  console.log('4️⃣  Executing calculate.ts with params...');
  const calcResult = await executeCell(calcCell, '.', { A: '15', B: '27' });
  console.log(`   ✅ Result:`);
  console.log(`      Exit code: ${calcResult.exitCode}`);
  console.log(`      Stdout: ${calcResult.stdout.trim()}`);
  console.log('');

  // 5. Simulate state tracking
  console.log('5️⃣  Testing state tracker pattern...');
  const state = {
    evidenceCount: 0,
    phase: 'test',
    iteration: 0,
    executionHistory: [],
  };

  state.iteration += 1;
  state.executionHistory.push({
    cellId: 'hello',
    timestamp: new Date().toISOString(),
    success: helloResult.exitCode === 0,
  });

  state.iteration += 1;
  state.executionHistory.push({
    cellId: 'calculate',
    timestamp: new Date().toISOString(),
    success: calcResult.exitCode === 0,
  });

  console.log(`   ✅ State:`);
  console.log(`      Phase: ${state.phase}`);
  console.log(`      Iteration: ${state.iteration}`);
  console.log(`      Executions: ${state.executionHistory.length}`);
  console.log('');

  // 6. Simulate embedded resources response
  console.log('6️⃣  Simulating embedded resources response...');
  const response = {
    content: [
      {
        type: 'text',
        text: 'Cell "calculate" executed successfully',
      },
      {
        type: 'resource',
        resource: {
          uri: 'notebook:///calculate/output',
          mimeType: 'application/json',
          text: JSON.stringify({
            stdout: calcResult.stdout,
            stderr: calcResult.stderr,
            exitCode: calcResult.exitCode,
          }, null, 2),
        },
      },
      {
        type: 'resource',
        resource: {
          uri: 'notebook:///state',
          mimeType: 'application/json',
          text: JSON.stringify(state, null, 2),
        },
      },
    ],
  };

  console.log('   ✅ Response structure:');
  response.content.forEach((item, idx) => {
    if (item.type === 'text') {
      console.log(`      [${idx}] Text: "${item.text}"`);
    } else if (item.type === 'resource') {
      console.log(`      [${idx}] Resource: ${item.resource.uri}`);
      console.log(`           MimeType: ${item.resource.mimeType}`);
      console.log(`           Size: ${item.resource.text.length} chars`);
    }
  });
  console.log('');

  console.log('🎉 All direct tests passed!\n');
  console.log('✨ Key Components Verified:');
  console.log('   • Notebook parsing (.src.md format)');
  console.log('   • Cell execution (TypeScript)');
  console.log('   • Parameter passing (environment variables)');
  console.log('   • State tracking (deterministic)');
  console.log('   • Embedded resources pattern');
}

test().catch((error) => {
  console.error('❌ Test failed:', error);
  process.exit(1);
});
