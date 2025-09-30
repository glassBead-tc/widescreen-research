// Core type definitions for research agent notebooks

export interface Cell {
  id: string;
  type: 'markdown' | 'code' | 'package.json';
  source: string;
  filename?: string;
}

export interface Notebook {
  cells: Cell[];
  language: 'javascript' | 'typescript';
}

export interface ExecutionResult {
  stdout: string;
  stderr: string;
  exitCode: number;
  error?: string;
}