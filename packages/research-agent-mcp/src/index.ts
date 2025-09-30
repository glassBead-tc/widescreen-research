#!/usr/bin/env node
// Research Agent MCP Runtime - Complete agent in ~400 LOC

import { readFileSync } from 'fs';
import { resolve } from 'path';
import { parseSrcmd } from './srcmd-parser.js';
import { createAgentServer, startServer } from './mcp-server.js';
import { createAgentClient } from './mcp-client.js';
import type { Notebook } from './types.js';

async function main() {
  // Configuration from environment
  const agentId = process.env.AGENT_ID || 'agent-1';
  const notebookPath = process.env.NOTEBOOK_PATH || './notebook.src.md';
  const workdir = process.env.WORKDIR || process.cwd();

  console.error(`[AGENT-${agentId}] Starting research agent...`);
  console.error(`[AGENT-${agentId}] Notebook: ${notebookPath}`);

  // 1. Parse notebook
  console.error(`[AGENT-${agentId}] Parsing notebook...`);
  const notebookContent = readFileSync(resolve(workdir, notebookPath), 'utf8');
  const notebook = parseSrcmd(notebookContent);
  console.error(`[AGENT-${agentId}] Loaded ${notebook.cells.length} cells`);

  // 2. Inject helper cells (state-tracker, retrieve, peers)
  injectHelperCells(notebook, agentId);

  // 3. Create MCP server (expose to others)
  console.error(`[AGENT-${agentId}] Creating MCP server...`);
  const server = await createAgentServer(notebook, workdir, agentId);

  // 4. Create MCP client (call others)
  console.error(`[AGENT-${agentId}] Creating MCP client...`);
  const client = await createAgentClient({
    agentId,
    externalServers: getExternalServers(),
    peerAgentUrls: getPeerAgentUrls()
  });

  // 5. Make client available globally for notebook cells
  (global as any).mcpClient = client;
  console.error(`[AGENT-${agentId}] MCP client available globally`);

  // 6. Start MCP server
  console.error(`[AGENT-${agentId}] Starting MCP server on stdio...`);
  await startServer(server);
}

/**
 * Inject helper cells into notebook
 */
function injectHelperCells(notebook: Notebook, agentId: string): void {
  // Inject state tracker at beginning
  notebook.cells.unshift({
    id: 'state-tracker',
    type: 'code',
    filename: 'state-tracker.ts',
    source: generateStateTrackerCode()
  });

  // Inject helpers
  notebook.cells.unshift({
    id: 'helpers',
    type: 'code',
    filename: 'helpers.ts',
    source: generateHelpersCode(agentId)
  });

  console.error(`[AGENT-${agentId}] Injected ${2} helper cells`);
}

/**
 * Generate state tracker code
 */
function generateStateTrackerCode(): string {
  return `
// State Tracker - Deterministic state management only

export interface AgentState {
  evidenceCount: number;
  phase: string;
  iteration: number;
  gates: Record<string, boolean>;
}

export const state: AgentState = {
  evidenceCount: 0,
  phase: 'observe',
  iteration: 0,
  gates: {}
};

// Deterministic state mutations only
export function recordEvidence(count: number): void {
  state.evidenceCount += count;
  console.log('[STATE] Evidence count:', state.evidenceCount);
}

export function setPhase(phase: string): void {
  state.phase = phase;
  console.log('[STATE] Phase:', phase);
}

export function incrementIteration(): void {
  state.iteration += 1;
  console.log('[STATE] Iteration:', state.iteration);
}

// Gate checking - deterministic boolean logic only
export function checkGate(gateName: string): boolean {
  switch(gateName) {
    case 'evidence-gathering':
      return state.evidenceCount >= 20;
    case 'synthesis':
      return state.evidenceCount >= 20 && state.phase === 'synthesis';
    case 'iteration':
      return state.iteration < 5;
    default:
      return false;
  }
}

console.log('[STATE] State tracker initialized');
`.trim();
}

/**
 * Generate helpers code
 */
function generateHelpersCode(agentId: string): string {
  return `
// Auto-injected helpers for agent: ${agentId}

// Global MCP client (set by runtime)
declare const mcpClient: any;

// Smart retrieval: Exa for discovery → Firecrawl for extraction
export async function retrieve(query: string, url?: string): Promise<any> {
  if (url) {
    console.log('[RETRIEVE] Using Firecrawl (known URL)');
    return await mcpClient.callTool({
      name: 'firecrawl_scrape',
      arguments: { url, formats: ['markdown'] }
    });
  }

  console.log('[RETRIEVE] Step 1: Discovery with Exa');
  const results = await mcpClient.callTool({
    name: 'web_search_exa',
    arguments: { query, numResults: 5 }
  });

  console.log('[RETRIEVE] Step 2: Extract with Firecrawl');
  const contents = [];
  for (const result of results.results || []) {
    const content = await mcpClient.callTool({
      name: 'firecrawl_scrape',
      arguments: { url: result.url, formats: ['markdown'] }
    });
    contents.push({ url: result.url, content });
  }

  return contents;
}

// Discover peer agents
export function discoverPeers(): string[] {
  const urls = (process.env.PEER_AGENT_URLS || '').split(',');
  return urls.filter(u => u.trim());
}

// Call peer agent tool
export async function callPeer(agentId: string, toolName: string, args: any): Promise<any> {
  return await mcpClient.callTool({
    name: \`\${agentId}.\${toolName}\`,
    arguments: args
  });
}

console.log('[HELPERS] Loaded for ${agentId}');
`.trim();
}

/**
 * Get external MCP servers from environment
 */
function getExternalServers() {
  // Parse from env or use defaults
  return [
    // These would be configurable via env vars
    // For now, return empty - will be added as needed
  ];
}

/**
 * Get peer agent URLs from environment
 */
function getPeerAgentUrls(): string[] {
  const urls = process.env.PEER_AGENT_URLS || '';
  return urls.split(',').filter(u => u.trim());
}

// Run
main().catch((error) => {
  console.error('[AGENT] Fatal error:', error);
  process.exit(1);
});