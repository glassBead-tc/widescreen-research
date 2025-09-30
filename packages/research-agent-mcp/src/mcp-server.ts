// MCP Server - Exposes notebook operations to other agents

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import type { Notebook, Cell } from './types.js';
import { executeCell } from './executor.js';
import { findCell } from './srcmd-parser.js';

export async function createAgentServer(
  notebook: Notebook,
  workdir: string,
  agentId: string = 'agent'
): Promise<Server> {
  const server = new Server(
    {
      name: `research-agent-${agentId}`,
      version: '1.0.0'
    },
    {
      capabilities: {
        tools: {},
        resources: {}
      }
    }
  );

  // Tool: Execute Cell
  server.setToolHandler(async (request) => {
    if (request.params.name === 'execute_cell') {
      const args = request.params.arguments as any;
      const cell = findCell(notebook, args.cellId);

      if (!cell) {
        return {
          tools: [],
          isError: true,
          content: [{ type: 'text', text: `Cell ${args.cellId} not found` }]
        };
      }

      const result = await executeCell(cell, workdir, args.params);

      return {
        tools: [],
        content: [
          {
            type: 'text',
            text: JSON.stringify(result, null, 2)
          }
        ]
      };
    }

    // Tool: Read Cell
    if (request.params.name === 'read_cell') {
      const args = request.params.arguments as any;
      const cell = findCell(notebook, args.cellId);

      return {
        tools: [],
        content: [
          {
            type: 'text',
            text: cell?.source || ''
          }
        ]
      };
    }

    // Tool: Write Cell (for collaborative editing)
    if (request.params.name === 'write_cell') {
      const args = request.params.arguments as any;
      const cell = findCell(notebook, args.cellId);

      if (cell) {
        cell.source = args.content;
      }

      return {
        tools: [],
        content: [{ type: 'text', text: 'Cell updated' }]
      };
    }

    // Tool: List Tools (for discovery)
    return {
      tools: [
        {
          name: 'execute_cell',
          description: 'Execute a notebook cell with optional parameters',
          inputSchema: {
            type: 'object',
            properties: {
              cellId: { type: 'string', description: 'Cell ID or filename' },
              params: { type: 'object', description: 'Environment variables for execution' }
            },
            required: ['cellId']
          }
        },
        {
          name: 'read_cell',
          description: 'Read cell source code (for other agents to access)',
          inputSchema: {
            type: 'object',
            properties: {
              cellId: { type: 'string', description: 'Cell ID or filename' }
            },
            required: ['cellId']
          }
        },
        {
          name: 'write_cell',
          description: 'Write content to a cell (for collaboration)',
          inputSchema: {
            type: 'object',
            properties: {
              cellId: { type: 'string', description: 'Cell ID or filename' },
              content: { type: 'string', description: 'New cell content' }
            },
            required: ['cellId', 'content']
          }
        }
      ]
    };
  });

  // Resources: Expose all cells as MCP resources
  server.setResourceHandler(async (request) => {
    if (request.method === 'resources/list') {
      return {
        resources: notebook.cells
          .filter(c => c.type === 'code')
          .map(c => ({
            uri: `notebook:///${c.id}`,
            name: c.filename || c.id,
            description: `Notebook cell: ${c.filename}`,
            mimeType: 'text/typescript'
          }))
      };
    }

    if (request.method === 'resources/read') {
      const cellId = request.params.uri.replace('notebook:///', '');
      const cell = findCell(notebook, cellId);

      return {
        contents: [
          {
            uri: request.params.uri,
            mimeType: 'text/typescript',
            text: cell?.source || ''
          }
        ]
      };
    }

    return { resources: [] };
  });

  return server;
}

export async function startServer(server: Server): Promise<void> {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('[AGENT] MCP server started on stdio');
}