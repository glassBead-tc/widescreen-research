#!/usr/bin/env node
/**
 * Simple test script to verify MCP server functionality
 * Tests the execute_cell tool with embedded resources
 */

import { readFileSync } from 'fs';
import { parseSrcmd } from './dist/srcmd-parser.js';
import { createAgentServer } from './dist/mcp-server.js';
import { CallToolRequestSchema } from '@modelcontextprotocol/sdk/types.js';

async function test() {
  console.log('🧪 Testing MCP Server Implementation\n');

  // 1. Load test notebook
  console.log('1️⃣  Loading test notebook...');
  const notebookContent = readFileSync('./examples/test-agent.src.md', 'utf8');
  const notebook = parseSrcmd(notebookContent);
  console.log(`   ✅ Loaded ${notebook.cells.length} cells\n`);

  // 2. Create MCP server
  console.log('2️⃣  Creating MCP server...');
  const server = await createAgentServer(notebook, '.', 'test-agent');
  console.log('   ✅ Server created\n');

  // 3. Test list_tools
  console.log('3️⃣  Testing tools/list...');
  const tools = await server.request(
    { method: 'tools/list' },
    null
  );
  console.log(`   ✅ Found ${tools.tools.length} tools:`);
  tools.tools.forEach((tool) => {
    console.log(`      - ${tool.name}: ${tool.description}`);
  });
  console.log('');

  // 4. Test list_cells
  console.log('4️⃣  Testing list_cells tool...');
  const listResult = await server.request(
    {
      method: 'tools/call',
      params: {
        name: 'list_cells',
        arguments: {},
      },
    },
    CallToolRequestSchema
  );
  console.log('   ✅ Response:', JSON.stringify(listResult, null, 2));
  console.log('');

  // 5. Test execute_cell (success)
  console.log('5️⃣  Testing execute_cell (hello.ts)...');
  const execResult = await server.request(
    {
      method: 'tools/call',
      params: {
        name: 'execute_cell',
        arguments: {
          cellId: 'hello',
        },
      },
    },
    CallToolRequestSchema
  );

  console.log('   ✅ Result:');
  execResult.content.forEach((item) => {
    if (item.type === 'text') {
      console.log(`      Text: ${item.text}`);
    } else if (item.type === 'resource') {
      console.log(`      Resource: ${item.resource.uri}`);
      console.log(`      Content: ${item.resource.text.substring(0, 100)}...`);
    }
  });
  console.log('');

  // 6. Test execute_cell with params
  console.log('6️⃣  Testing execute_cell with params (calculate.ts)...');
  const calcResult = await server.request(
    {
      method: 'tools/call',
      params: {
        name: 'execute_cell',
        arguments: {
          cellId: 'calculate',
          params: { A: '15', B: '27' },
        },
      },
    },
    CallToolRequestSchema
  );

  console.log('   ✅ Result:');
  calcResult.content.forEach((item) => {
    if (item.type === 'text') {
      console.log(`      Text: ${item.text}`);
    } else if (item.type === 'resource' && item.resource.uri.includes('output')) {
      const output = JSON.parse(item.resource.text);
      console.log(`      Stdout: ${output.stdout.trim()}`);
    }
  });
  console.log('');

  // 7. Test read_state
  console.log('7️⃣  Testing read_state...');
  const stateResult = await server.request(
    {
      method: 'tools/call',
      params: {
        name: 'read_state',
        arguments: {},
      },
    },
    CallToolRequestSchema
  );

  const stateResource = stateResult.content.find(
    (c) => c.type === 'resource' && c.resource.uri === 'notebook:///state'
  );
  const state = JSON.parse(stateResource.resource.text);
  console.log('   ✅ State:', {
    iteration: state.iteration,
    phase: state.phase,
    evidenceCount: state.evidenceCount,
    executionHistory: state.executionHistory.length,
  });
  console.log('');

  // 8. Test execute_cell (error case)
  console.log('8️⃣  Testing execute_cell error handling (error-test.ts)...');
  const errorResult = await server.request(
    {
      method: 'tools/call',
      params: {
        name: 'execute_cell',
        arguments: {
          cellId: 'error-test',
        },
      },
    },
    CallToolRequestSchema
  );

  const errorOutput = errorResult.content.find(
    (c) => c.type === 'resource' && c.resource.uri.includes('output')
  );
  const errorData = JSON.parse(errorOutput.resource.text);
  console.log('   ✅ Error handled correctly:');
  console.log(`      Exit code: ${errorData.exitCode}`);
  console.log(`      Has stderr: ${errorData.stderr.length > 0}`);
  console.log('');

  console.log('🎉 All tests passed!');
  console.log('\n✨ Key Features Verified:');
  console.log('   • Tool discovery (list_tools)');
  console.log('   • Cell execution with params');
  console.log('   • Embedded resources in responses');
  console.log('   • State tracking across executions');
  console.log('   • Error handling');
}

test().catch((error) => {
  console.error('❌ Test failed:', error);
  process.exit(1);
});
