import { spawn } from 'child_process';
import { writeFile, mkdir } from 'fs/promises';
import { join } from 'path';
import type { Cell, ExecutionResult } from './types.js';

export async function executeCell(
  cell: Cell,
  workdir: string,
  params?: Record<string, string>
): Promise<ExecutionResult> {
  if (cell.type !== 'code') {
    throw new Error(`Cannot execute ${cell.type} cell`);
  }

  const tempDir = join(workdir, '.agent-runtime');
  await mkdir(tempDir, { recursive: true });

  const cellPath = join(tempDir, cell.filename || `${cell.id}.ts`);
  await writeFile(cellPath, cell.source);

  const isTypeScript = cell.filename?.endsWith('.ts');
  const command = isTypeScript ? 'npx' : 'node';
  const args = isTypeScript ? ['tsx', cellPath] : [cellPath];

  return new Promise((resolve) => {
    const child = spawn(command, args, {
      cwd: workdir,
      env: { ...process.env, ...params }
    });

    let stdout = '';
    let stderr = '';

    child.stdout.on('data', (data) => { stdout += data.toString(); });
    child.stderr.on('data', (data) => { stderr += data.toString(); });

    child.on('close', (code) => {
      resolve({ stdout, stderr, exitCode: code || 0, error: code !== 0 ? stderr : undefined });
    });

    child.on('error', (err) => {
      resolve({ stdout, stderr, exitCode: 1, error: err.message });
    });
  });
}
