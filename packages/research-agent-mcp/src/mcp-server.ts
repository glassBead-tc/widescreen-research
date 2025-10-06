// MCP Server - Exposes notebook operations to other agents
// Implements embedded resources pattern (resources returned in tool responses)

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import type { Notebook, Cell, ExecutionResult } from './types.js';
import { executeCell } from './executor.js';
import { findCell } from './srcmd-parser.js';

/**
 * State tracker for the agent
 * Server tracks state, doesn't make decisions
 */
interface AgentState {
  evidenceCount: number;
  phase: string;
  iteration: number;
  executionHistory: Array<{
    cellId: string;
    timestamp: string;
    success: boolean;
  }>;
  metadata: Record<string, any>;
}

/**
 * Create MCP server for a notebook agent
 */
export async function createAgentServer(
  notebook: Notebook,
  workdir: string,
  agentId: string = 'agent'
): Promise<Server> {
  // Initialize state tracker
  const state: AgentState = {
    evidenceCount: 0,
    phase: 'init',
    iteration: 0,
    executionHistory: [],
    metadata: {},
  };

  const server = new Server(
    {
      name: `research-agent-${agentId}`,
      version: '1.0.0',
    },
    {
      capabilities: {
        tools: {},
      },
    }
  );

  /**
   * List available tools
   */
  server.setRequestHandler(ListToolsRequestSchema, async () => ({
    tools: [
      {
        name: 'execute_cell',
        description: 'Execute a notebook cell and return results with embedded resources',
        inputSchema: {
          type: 'object',
          properties: {
            cellId: {
              type: 'string',
              description: 'Cell ID or filename to execute',
            },
            params: {
              type: 'object',
              description: 'Optional parameters passed as environment variables',
            },
          },
          required: ['cellId'],
        },
      },
      {
        name: 'read_state',
        description: 'Read the current agent state (evidence count, phase, etc.)',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'list_cells',
        description: 'List all available cells in the notebook',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'read_cell',
        description: 'Read cell source code without executing',
        inputSchema: {
          type: 'object',
          properties: {
            cellId: {
              type: 'string',
              description: 'Cell ID or filename',
            },
          },
          required: ['cellId'],
        },
      },
    ],
  }));

  /**
   * Handle tool calls
   */
  server.setRequestHandler(CallToolRequestSchema, async (request) => {
    const { name, arguments: args } = request.params;

    try {
      switch (name) {
        case 'execute_cell':
          return await handleExecuteCell(args as any, notebook, workdir, state);

        case 'read_state':
          return handleReadState(state);

        case 'list_cells':
          return handleListCells(notebook);

        case 'read_cell':
          return handleReadCell(args as any, notebook);

        default:
          throw new Error(`Unknown tool: ${name}`);
      }
    } catch (error) {
      return {
        content: [
          {
            type: 'text',
            text: `Error: ${error instanceof Error ? error.message : String(error)}`,
          },
        ],
        isError: true,
      };
    }
  });

  return server;
}

/**
 * Handle execute_cell tool call
 * Returns execution results with embedded resources (output + state)
 */
async function handleExecuteCell(
  args: { cellId: string; params?: Record<string, string> },
  notebook: Notebook,
  workdir: string,
  state: AgentState
): Promise<any> {
  const { cellId, params } = args;

  // Find cell
  const cell = findCell(notebook, cellId);
  if (!cell) {
    throw new Error(`Cell not found: ${cellId}`);
  }

  if (cell.type !== 'code') {
    throw new Error(`Cannot execute ${cell.type} cell`);
  }

  // Execute cell
  const startTime = Date.now();
  const result = await executeCell(cell, workdir, params);
  const endTime = Date.now();

  // Update state (deterministic tracking only)
  const success = result.exitCode === 0;
  state.executionHistory.push({
    cellId,
    timestamp: new Date().toISOString(),
    success,
  });

  if (success) {
    state.iteration += 1;
  }

  // Return with embedded resources
  return {
    content: [
      {
        type: 'text',
        text: success
          ? `Cell "${cellId}" executed successfully in ${endTime - startTime}ms`
          : `Cell "${cellId}" execution failed`,
      },
      {
        type: 'resource',
        resource: {
          uri: `notebook:///${cellId}/output`,
          mimeType: 'application/json',
          text: JSON.stringify(
            {
              stdout: result.stdout,
              stderr: result.stderr,
              exitCode: result.exitCode,
              executionTime: endTime - startTime,
            },
            null,
            2
          ),
        },
      },
      {
        type: 'resource',
        resource: {
          uri: 'notebook:///state',
          mimeType: 'application/json',
          text: JSON.stringify(
            {
              evidenceCount: state.evidenceCount,
              phase: state.phase,
              iteration: state.iteration,
              lastExecution: {
                cellId,
                success,
                timestamp: new Date().toISOString(),
              },
            },
            null,
            2
          ),
        },
      },
    ],
  };
}

/**
 * Handle read_state tool call
 * Returns current agent state
 */
function handleReadState(state: AgentState): any {
  return {
    content: [
      {
        type: 'text',
        text: `Agent state: Phase "${state.phase}", Iteration ${state.iteration}, Evidence count: ${state.evidenceCount}`,
      },
      {
        type: 'resource',
        resource: {
          uri: 'notebook:///state',
          mimeType: 'application/json',
          text: JSON.stringify(
            {
              evidenceCount: state.evidenceCount,
              phase: state.phase,
              iteration: state.iteration,
              executionHistory: state.executionHistory,
              metadata: state.metadata,
            },
            null,
            2
          ),
        },
      },
    ],
  };
}

/**
 * Handle list_cells tool call
 * Returns list of available cells
 */
function handleListCells(notebook: Notebook): any {
  const cells = notebook.cells
    .filter((c) => c.type === 'code')
    .map((c) => ({
      id: c.id,
      filename: c.filename,
      type: c.type,
    }));

  return {
    content: [
      {
        type: 'text',
        text: `Notebook contains ${cells.length} executable cells`,
      },
      {
        type: 'resource',
        resource: {
          uri: 'notebook:///cells',
          mimeType: 'application/json',
          text: JSON.stringify({ cells, total: cells.length }, null, 2),
        },
      },
    ],
  };
}

/**
 * Handle read_cell tool call
 * Returns cell source code
 */
function handleReadCell(args: { cellId: string }, notebook: Notebook): any {
  const cell = findCell(notebook, args.cellId);

  if (!cell) {
    throw new Error(`Cell not found: ${args.cellId}`);
  }

  return {
    content: [
      {
        type: 'text',
        text: `Cell "${args.cellId}" source (${cell.source.split('\n').length} lines)`,
      },
      {
        type: 'resource',
        resource: {
          uri: `notebook:///${args.cellId}/source`,
          mimeType:
            cell.filename?.endsWith('.ts') ? 'text/typescript' : 'text/javascript',
          text: cell.source,
        },
      },
    ],
  };
}

/**
 * Start MCP server on stdio transport
 */
export async function startServer(server: Server): Promise<void> {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('[MCP-SERVER] Started on stdio transport');
}
