// Cell executor - runs TypeScript/JavaScript cells

import { spawn } from 'child_process';
import { writeFile, mkdir } from 'fs/promises';
import { join } from 'path';
import type { Cell, ExecutionResult } from './types.js';

/**
 * Execute a TypeScript or JavaScript cell
 */
export async function executeCell(
  cell: Cell,
  workdir: string,
  params?: Record<string, string>
): Promise<ExecutionResult> {
  if (cell.type !== 'code') {
    throw new Error(`Cannot execute ${cell.type} cell`);
  }

  // Create temp directory for execution
  const tempDir = join(workdir, '.agent-runtime');
  await mkdir(tempDir, { recursive: true });

  // Write cell to file
  const cellPath = join(tempDir, cell.filename || `${cell.id}.ts`);
  await writeFile(cellPath, cell.source);

  // Execute with tsx (TypeScript) or node (JavaScript)
  const isTypeScript = cell.filename?.endsWith('.ts') || cell.filename?.endsWith('.tsx');
  const command = isTypeScript ? 'npx' : 'node';
  const args = isTypeScript ? ['tsx', cellPath] : [cellPath];

  return new Promise((resolve) => {
    const child = spawn(command, args, {
      cwd: workdir,
      env: { ...process.env, ...params }
    });

    let stdout = '';
    let stderr = '';

    child.stdout.on('data', (data) => {
      stdout += data.toString();
    });

    child.stderr.on('data', (data) => {
      stderr += data.toString();
    });

    child.on('close', (code) => {
      resolve({
        stdout,
        stderr,
        exitCode: code || 0,
        error: code !== 0 ? stderr : undefined
      });
    });

    child.on('error', (err) => {
      resolve({
        stdout,
        stderr,
        exitCode: 1,
        error: err.message
      });
    });
  });
}

/**
 * Execute multiple cells in sequence
 */
export async function executeCells(
  cells: Cell[],
  workdir: string,
  params?: Record<string, string>
): Promise<ExecutionResult[]> {
  const results: ExecutionResult[] = [];

  for (const cell of cells) {
    if (cell.type === 'code') {
      const result = await executeCell(cell, workdir, params);
      results.push(result);

      // Stop on error
      if (result.exitCode !== 0) {
        break;
      }
    }
  }

  return results;
}