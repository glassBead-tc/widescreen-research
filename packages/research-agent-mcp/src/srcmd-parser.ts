// Minimal .src.md parser

import type { Cell, Notebook } from './types.js';

export function parseSrcmd(content: string): Notebook {
  const cells: Cell[] = [];

  const langMatch = content.match(/<!--\s*srcbook:\s*{\s*"language"\s*:\s*"(\w+)"\s*}/);
  const language = (langMatch?.[1] || 'typescript') as 'javascript' | 'typescript';

  const cellPattern = /######\s+([^\n]+)\n\n```(?:typescript|javascript|json)\n([\s\S]*?)```/g;

  let match;
  while ((match = cellPattern.exec(content)) !== null) {
    const filename = match[1].trim();
    const source = match[2].trim();

    cells.push({
      id: filename.replace(/\.(ts|js|json)$/, ''),
      type: filename === 'package.json' ? 'package.json' : 'code',
      source,
      filename
    });
  }

  return { cells, language };
}

export function findCell(notebook: Notebook, idOrFilename: string): Cell | undefined {
  return notebook.cells.find(
    c => c.id === idOrFilename || c.filename === idOrFilename
  );
}