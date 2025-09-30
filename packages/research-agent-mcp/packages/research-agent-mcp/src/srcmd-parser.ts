import type { Cell, Notebook } from './types.js';

export function parseSrcmd(content: string): Notebook {
  const cells: Cell[] = [];
  const langMatch = content.match(/<!--\s*srcbook:\s*{\s*"language"\s*:\s*"(\w+)"\s*}/);
  const language = (langMatch?.[1] || 'typescript') as 'javascript' | 'typescript';

  const cellPattern = /######\s+([^\n]+)\n\n```(?:typescript|javascript|json)\n([\s\S]*?)```/g;
  let match;
  while ((match = cellPattern.exec(content)) !== null) {
    cells.push({
      id: match[1].trim().replace(/\.(ts|js|json)$/, ''),
      type: match[1].trim() === 'package.json' ? 'package.json' : 'code',
      source: match[2].trim(),
      filename: match[1].trim()
    });
  }
  return { cells, language };
}

export function findCell(notebook: Notebook, id: string): Cell | undefined {
  return notebook.cells.find(c => c.id === id || c.filename === id);
}
