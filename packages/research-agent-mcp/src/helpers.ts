// Helper functions - Auto-injected into notebook context

import type { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { callTool } from './mcp-client.js';

/**
 * Smart retrieval: Exa for discovery → Firecrawl for extraction
 * Auto-injected into every agent notebook
 */
export async function retrieve(
  client: Client,
  query: string,
  url?: string
): Promise<any> {
  if (url) {
    // Have URL → use Firecrawl directly
    console.error(`[RETRIEVE] Using Firecrawl (known URL: ${url})`);
    return await callTool(client, 'firecrawl_scrape', {
      url,
      formats: ['markdown']
    });
  }

  // No URL → Discovery with Exa first
  console.error(`[RETRIEVE] Step 1: Discovery with Exa (query: ${query})`);
  const searchResults = await callTool(client, 'web_search_exa', {
    query,
    numResults: 5
  });

  // Step 2: Extract content with Firecrawl
  console.error(`[RETRIEVE] Step 2: Extraction with Firecrawl (${searchResults.results?.length || 0} URLs)`);
  const contents = [];

  for (const result of searchResults.results || []) {
    try {
      const content = await callTool(client, 'firecrawl_scrape', {
        url: result.url,
        formats: ['markdown']
      });
      contents.push({
        url: result.url,
        title: result.title,
        content
      });
    } catch (error) {
      console.error(`[RETRIEVE] Failed to scrape ${result.url}:`, error);
    }
  }

  return contents;
}

/**
 * Discover peer agents via environment variables
 * Auto-injected into every agent notebook
 */
export function discoverPeers(): string[] {
  const peerUrls = process.env.PEER_AGENT_URLS || '';
  return peerUrls.split(',').filter(u => u.trim());
}

/**
 * Call peer agent tool
 * Auto-injected into every agent notebook
 */
export async function callPeer(
  client: Client,
  agentId: string,
  toolName: string,
  args: any
): Promise<any> {
  return await callTool(client, `${agentId}.${toolName}`, args);
}

/**
 * Generate helper code to inject into notebook
 */
export function generateHelperCode(agentId: string): string {
  return `
// Auto-injected helpers for agent: ${agentId}

// Global MCP client (set by runtime)
declare const mcpClient: any;

// Smart retrieval: Exa → Firecrawl
export async function retrieve(query: string, url?: string) {
  ${retrieve.toString()}
}

// Discover peer agents
export function discoverPeers(): string[] {
  ${discoverPeers.toString()}
}

// Call peer agent
export async function callPeer(agentId: string, toolName: string, args: any) {
  ${callPeer.toString()}
}

console.log('[HELPERS] Loaded for agent: ${agentId}');
`.trim();
}