#!/usr/bin/env node
import { readFileSync } from 'fs';
import { parseSrcmd } from './srcmd-parser.js';

async function main() {
  const agentId = process.env.AGENT_ID || 'agent-1';
  const notebookPath = process.env.NOTEBOOK_PATH || './notebook.src.md';

  console.error(`[${agentId}] Starting...`);
  console.error(`[${agentId}] Notebook: ${notebookPath}`);

  const notebook = parseSrcmd(readFileSync(notebookPath, 'utf8'));
  console.error(`[${agentId}] Loaded ${notebook.cells.length} cells`);

  // TODO: Implement MCP server/client startup
  console.error(`[${agentId}] Ready (minimal runtime)`);
}

main().catch(err => {
  console.error('Fatal:', err);
  process.exit(1);
});
