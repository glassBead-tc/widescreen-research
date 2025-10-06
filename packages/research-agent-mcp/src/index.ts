#!/usr/bin/env node
// Research Agent MCP Runtime - Complete agent in ~400 LOC

import { readFileSync } from 'fs';
import { resolve } from 'path';
import { parseSrcmd } from './srcmd-parser.js';
import { createAgentServer, startServer } from './mcp-server.js';
import { createAgentClient, type NotebookMCPClient } from './mcp-client.js';
import { generateHelpersCode, HelperInjector } from './helpers.js';
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
  const helperInjector = new HelperInjector(client);
  (global as any).mcpClient = helperInjector.getGlobalClient();
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
 * Get external MCP servers from environment
 */
function getExternalServers() {
  const serversJson = process.env.EXTERNAL_MCP_SERVERS;

  if (serversJson) {
    try {
      return JSON.parse(serversJson);
    } catch (error) {
      console.error('[AGENT] Failed to parse EXTERNAL_MCP_SERVERS:', error);
    }
  }

  // Default: no external servers
  // Users can add them via EXTERNAL_MCP_SERVERS env var
  return [];
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
