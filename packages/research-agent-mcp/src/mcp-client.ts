// MCP Client - Calls other agents and external MCP servers

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';

export interface MCPClientConfig {
  agentId: string;
  externalServers?: Array<{ name: string; command: string; args: string[] }>;
  peerAgentUrls?: string[];
}

/**
 * Create MCP client that can call both external services and peer agents
 */
export async function createAgentClient(config: MCPClientConfig): Promise<Client> {
  const client = new Client(
    {
      name: `research-client-${config.agentId}`,
      version: '1.0.0'
    },
    {
      capabilities: {}
    }
  );

  // Connect to external MCP servers (arxiv, exa, firecrawl, etc.)
  if (config.externalServers) {
    for (const server of config.externalServers) {
      try {
        const transport = new StdioClientTransport({
          command: server.command,
          args: server.args
        });
        await client.connect(transport);
        console.error(`[CLIENT] Connected to external server: ${server.name}`);
      } catch (error) {
        console.error(`[CLIENT] Failed to connect to ${server.name}:`, error);
      }
    }
  }

  // TODO: Connect to peer agents via HTTP transport
  // This would require HTTP transport implementation
  if (config.peerAgentUrls) {
    console.error(`[CLIENT] Peer agent connections not yet implemented`);
    // Future: HTTP transport for peer-to-peer MCP
  }

  return client;
}

/**
 * Helper: Call a tool on the MCP client
 */
export async function callTool(
  client: Client,
  toolName: string,
  args: any
): Promise<any> {
  try {
    const result = await client.callTool({
      name: toolName,
      arguments: args
    });

    return result;
  } catch (error) {
    console.error(`[CLIENT] Tool call failed: ${toolName}`, error);
    throw error;
  }
}

/**
 * Helper: List available tools from all connected servers
 */
export async function listAllTools(client: Client): Promise<any[]> {
  try {
    const result = await client.listTools();
    return result.tools || [];
  } catch (error) {
    console.error('[CLIENT] Failed to list tools:', error);
    return [];
  }
}