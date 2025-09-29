#!/usr/bin/env node
// Standalone MCP Server Test - Tests widescreen-research MCP server

import { spawn } from 'child_process';
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

async function testServer() {
  console.log('🧪 Testing Widescreen MCP Server\n');

  // Create client
  const client = new Client({
    name: 'test-client',
    version: '1.0.0'
  }, { capabilities: {} });

  // Create transport
  const transport = new StdioClientTransport({
    command: 'go',
    args: ['run', './cmd/widescreen-research-mcp'],
    env: {
      GOOGLE_CLOUD_PROJECT: 'widescreen-researcher',
      GCP_REGION: 'us-central1',
      EXA_API_KEY: '04d84ad6-726b-450d-9274-4050b08ab052',
      LOG_LEVEL: 'info'
    }
  });

  try {
    // Connect
    console.log('📡 Connecting to server...');
    await client.connect(transport);
    console.log('✅ Connected!\n');

    // List tools
    console.log('🔍 Discovering tools...');
    const tools = await client.listTools();
    console.log(`✅ Found ${tools.tools.length} tools:`);
    tools.tools.forEach(t => console.log(`  - ${t.name}: ${t.description}`));

    // Call a tool
    console.log('\n⚙️  Testing tool execution...');
    const result = await client.callTool({
      name: 'widescreen-research',
      arguments: { operation: 'start', query: 'Test query' }
    });
    console.log('✅ Tool executed');
    console.log('Result:', JSON.stringify(result, null, 2).substring(0, 300));

    // Cleanup
    await client.close();
    console.log('\n✅ All tests passed!');

  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
    process.exit(1);
  }
}

testServer();
