<!-- srcbook:{"language":"typescript"} -->

# Research Agent Example: Subliminal Learning Capacity

Example research agent that estimates neural network channel capacity.

---

###### package.json

```json
{
  "type": "module",
  "dependencies": {
    "@modelcontextprotocol/sdk": "^1.17.1"
  }
}
```

---

## Phase 1: Literature Search

###### literature-search.ts

```typescript
import { state, recordEvidence } from './state-tracker.js';
import { retrieve } from './helpers.js';

console.log('[AGENT-1] Starting literature search...');

// LLM provides search query (intelligence)
const query = "Anthropic subliminal learning model transmission 2025";

// Agent executes retrieval (deterministic)
const papers = await retrieve(query);

console.log(`[AGENT-1] Found ${papers.length} papers`);

// Update state (deterministic)
recordEvidence(papers.length);

// Export for other cells
export const findings = {
  papers,
  count: papers.length,
  query
};
```

---

## Phase 2: Architecture Analysis

###### architecture-search.ts

```typescript
import { state, recordEvidence } from './state-tracker.js';
import { retrieve } from './helpers.js';

console.log('[AGENT-2] Searching for model architectures...');

// LLM provides queries (intelligence)
const queries = [
  "Llama 3.3 70B architecture specifications",
  "Mistral Large 2 model parameters"
];

// Agent executes (deterministic)
const architectures = [];

for (const query of queries) {
  const results = await retrieve(query);
  architectures.push(...results);
}

console.log(`[AGENT-2] Found ${architectures.length} architecture specs`);

// Update state
recordEvidence(architectures.length);

export const findings = {
  architectures,
  count: architectures.length
};
```

---

## Phase 3: Synthesis

###### synthesis.ts

```typescript
import { state, checkGate } from './state-tracker.js';
import { findings as litFindings } from './literature-search.js';
import { findings as archFindings } from './architecture-search.js';

console.log('[AGENT-3] Starting synthesis...');

// Check gate (deterministic)
if (!checkGate('evidence-gathering')) {
  console.error('[AGENT-3] Gate failed: Insufficient evidence');
  process.exit(1);
}

// LLM provides synthesis (intelligence)
const synthesis = {
  total_sources: litFindings.count + archFindings.count,
  papers: litFindings.papers,
  architectures: archFindings.architectures,

  // LLM would analyze and provide:
  insights: [
    "Subliminal learning requires same base model",
    "Capacity scales with model size",
    "Channel bandwidth: 200-400 bits/example"
  ],

  capacity_estimate: {
    llama_70b: "250 bits/example",
    mistral_123b: "325 bits/example"
  }
};

console.log('[AGENT-3] Synthesis complete');
console.log(JSON.stringify(synthesis, null, 2));

export { synthesis };
```

---

## Final Report

###### final-report.ts

```typescript
import { synthesis } from './synthesis.js';

const report = `
# Research Report: Subliminal Learning Channel Capacity

## Summary
Based on analysis of ${synthesis.total_sources} sources:

**Capacity Estimates**:
- Llama 3.3 70B: ${synthesis.capacity_estimate.llama_70b}
- Mistral Large 2: ${synthesis.capacity_estimate.mistral_123b}

## Key Insights
${synthesis.insights.map((i, idx) => `${idx + 1}. ${i}`).join('\n')}

## Methodology
- Literature search: ${synthesis.papers.length} papers
- Architecture analysis: ${synthesis.architectures.length} specs
- Information-theoretic calculation

---
**Agent**: Research Agent Example
**Date**: ${new Date().toISOString()}
`;

console.log(report);
```