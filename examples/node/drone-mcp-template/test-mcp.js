// Minimal MCP server test
import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';

async function testMcp() {
  try {
    console.log('Creating server...');
    const server = new Server({
      name: 'test-server',
      version: '1.0.0'
    });

    console.log('Setting up handlers...');
    server.setRequestHandler('initialize', async () => {
      return {
        protocolVersion: '2024-11-05',
        capabilities: {},
        serverInfo: {
          name: 'test-server',
          version: '1.0.0'
        }
      };
    });

    console.log('Creating transport...');
    const transport = new StdioServerTransport();
    
    console.log('Connecting...');
    await server.connect(transport);
    
    console.log('MCP server started successfully!');
  } catch (error) {
    console.error('Error:', error.message);
    console.error('Stack:', error.stack);
  }
}

testMcp(); 