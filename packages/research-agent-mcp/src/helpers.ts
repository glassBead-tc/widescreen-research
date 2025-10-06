// Helper functions - Auto-injected into notebook context
// These helpers are available in all notebook cells

import type { NotebookMCPClient } from './mcp-client.js';

/**
 * Smart retrieval: Exa for discovery → Firecrawl for extraction
 * This is the pattern that gets injected into cells
 */
export function generateRetrieveHelper(): string {
  return `
async function retrieve(query, url) {
  if (url) {
    // Have URL → use Firecrawl directly
    console.log('[RETRIEVE] Using Firecrawl (known URL)');
    return await mcpClient.callTool({
      server: 'firecrawl',
      tool: 'firecrawl_scrape',
      arguments: { url, formats: ['markdown'] }
    });
  }

  // No URL → Discovery with Exa first
  console.log('[RETRIEVE] Step 1: Discovery with Exa');
  const searchResults = await mcpClient.callTool({
    server: 'exa',
    tool: 'web_search_exa',
    arguments: { query, numResults: 5 }
  });

  // Step 2: Extract content with Firecrawl
  console.log('[RETRIEVE] Step 2: Extraction with Firecrawl');
  const contents = [];

  for (const result of searchResults.results || []) {
    try {
      const content = await mcpClient.callTool({
        server: 'firecrawl',
        tool: 'firecrawl_scrape',
        arguments: { url: result.url, formats: ['markdown'] }
      });
      contents.push({
        url: result.url,
        title: result.title,
        content
      });
    } catch (error) {
      console.error('[RETRIEVE] Failed to scrape ' + result.url + ':', error);
    }
  }

  return contents;
}
`.trim();
}

/**
 * Discover peer agents via environment variables
 */
export function generateDiscoverPeersHelper(): string {
  return `
function discoverPeers() {
  const peerUrls = process.env.PEER_AGENT_URLS || '';
  return peerUrls.split(',').filter(u => u.trim());
}
`.trim();
}

/**
 * Call peer agent tool
 */
export function generateCallPeerHelper(): string {
  return `
async function callPeer(agentId, toolName, args) {
  return await mcpClient.callTool({
    server: agentId,
    tool: toolName,
    arguments: args
  });
}
`.trim();
}

/**
 * Generate complete helper code to inject into notebook cells
 */
export function generateHelpersCode(agentId: string): string {
  return `
// Auto-injected helpers for agent: ${agentId}

// Global MCP client (set by runtime)
declare const mcpClient: any;

${generateRetrieveHelper()}

${generateDiscoverPeersHelper()}

${generateCallPeerHelper()}

console.log('[HELPERS] Loaded for agent: ${agentId}');
`.trim();
}

/**
 * Helper utilities for working with MCP client
 */
export class HelperInjector {
  constructor(private client: NotebookMCPClient) {}

  /**
   * Get the global mcpClient object to inject into cells
   */
  getGlobalClient(): any {
    return {
      callTool: async (params: {
        server: string;
        tool: string;
        arguments?: Record<string, unknown>;
      }) => {
        return await this.client.callTool(params);
      },
      listTools: async (server: string) => {
        return await this.client.listTools(server);
      },
      getConnectedServers: () => {
        return this.client.getConnectedServers();
      },
    };
  }
}
