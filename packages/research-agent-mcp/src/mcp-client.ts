// MCP Client - Connect to external MCP servers and peer agents
// Enables notebook cells to call tools from other MCP servers

import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
import type { Transport } from '@modelcontextprotocol/sdk/shared/transport.js';

/**
 * Server configuration
 */
interface ServerConfig {
  name: string;
  transport: 'stdio' | 'http';
  command?: string;
  args?: string[];
  url?: string;
  env?: Record<string, string>;
}

/**
 * MCP Client for connecting to external servers and peer agents
 */
export class NotebookMCPClient {
  private clients: Map<string, Client> = new Map();
  private configs: Map<string, ServerConfig> = new Map();

  constructor(private agentId: string) {}

  /**
   * Connect to a stdio MCP server (e.g., arxiv, exa, firecrawl)
   */
  async connectStdio(config: ServerConfig): Promise<void> {
    if (config.transport !== 'stdio') {
      throw new Error('connectStdio requires transport: "stdio"');
    }

    if (!config.command || !config.args) {
      throw new Error('stdio transport requires command and args');
    }

    console.error(`[CLIENT] Connecting to ${config.name} via stdio...`);

    try {
      // Create transport
      const envVars: Record<string, string> = {};
      if (process.env) {
        for (const [key, value] of Object.entries(process.env)) {
          if (value !== undefined) {
            envVars[key] = value;
          }
        }
      }
      if (config.env) {
        Object.assign(envVars, config.env);
      }

      const transport = new StdioClientTransport({
        command: config.command,
        args: config.args,
        env: envVars,
      });

      // Create client
      const client = new Client(
        {
          name: `${this.agentId}-client`,
          version: '1.0.0',
        },
        {
          capabilities: {},
        }
      );

      // Connect
      await client.connect(transport);

      // Store client and config
      this.clients.set(config.name, client);
      this.configs.set(config.name, config);

      console.error(`[CLIENT] ✅ Connected to ${config.name}`);
    } catch (error) {
      console.error(`[CLIENT] ❌ Failed to connect to ${config.name}:`, error);
      throw error;
    }
  }

  /**
   * Connect to an HTTP MCP server (peer agents)
   */
  async connectHTTP(config: ServerConfig): Promise<void> {
    if (config.transport !== 'http') {
      throw new Error('connectHTTP requires transport: "http"');
    }

    if (!config.url) {
      throw new Error('http transport requires url');
    }

    console.error(`[CLIENT] HTTP connections not yet implemented`);
    console.error(`[CLIENT] Peer agent connections will be added in Phase 5`);

    // Store config for later
    this.configs.set(config.name, config);
  }

  /**
   * Call a tool on an external server
   */
  async callTool(params: {
    server: string;
    tool: string;
    arguments?: Record<string, unknown>;
  }): Promise<any> {
    const { server, tool, arguments: args } = params;

    const client = this.clients.get(server);
    if (!client) {
      throw new Error(
        `Not connected to server: ${server}. Available: ${Array.from(this.clients.keys()).join(', ')}`
      );
    }

    try {
      console.error(`[CLIENT] Calling ${server}.${tool}...`);
      const result = await client.callTool({ name: tool, arguments: args || {} });
      console.error(`[CLIENT] ✅ ${server}.${tool} succeeded`);
      return result;
    } catch (error) {
      console.error(`[CLIENT] ❌ ${server}.${tool} failed:`, error);
      throw error;
    }
  }

  /**
   * List available tools from a server
   */
  async listTools(server: string): Promise<any> {
    const client = this.clients.get(server);
    if (!client) {
      throw new Error(`Not connected to server: ${server}`);
    }

    try {
      const result = await client.listTools();
      return result;
    } catch (error) {
      console.error(`[CLIENT] Failed to list tools from ${server}:`, error);
      throw error;
    }
  }

  /**
   * Get list of connected servers
   */
  getConnectedServers(): string[] {
    return Array.from(this.clients.keys());
  }

  /**
   * Disconnect from a server
   */
  async disconnect(server: string): Promise<void> {
    const client = this.clients.get(server);
    if (client) {
      await client.close();
      this.clients.delete(server);
      console.error(`[CLIENT] Disconnected from ${server}`);
    }
  }

  /**
   * Disconnect from all servers
   */
  async disconnectAll(): Promise<void> {
    const servers = Array.from(this.clients.keys());
    for (const server of servers) {
      await this.disconnect(server);
    }
  }
}

/**
 * Create and configure MCP client for a notebook agent
 */
export async function createAgentClient(config: {
  agentId: string;
  externalServers?: ServerConfig[];
  peerAgentUrls?: string[];
}): Promise<NotebookMCPClient> {
  const client = new NotebookMCPClient(config.agentId);

  // Connect to external servers (stdio)
  if (config.externalServers && config.externalServers.length > 0) {
    for (const serverConfig of config.externalServers) {
      if (serverConfig.transport === 'stdio') {
        await client.connectStdio(serverConfig);
      } else if (serverConfig.transport === 'http') {
        await client.connectHTTP(serverConfig);
      }
    }
  } else {
    console.error('[CLIENT] No external servers configured');
  }

  // Connect to peer agents (HTTP) - Phase 5
  if (config.peerAgentUrls && config.peerAgentUrls.length > 0) {
    console.error('[CLIENT] Peer agent connections not yet implemented');
  }

  return client;
}
