# Validate State Tracker Pattern

## Description

Validates that MCP server operations follow the "state tracker only" principle. The MCP server must NEVER provide intelligence - it only validates, stores state, and returns progress metadata. All intelligence comes from the LLM client.

## Core Principle

> **The LLM is the intelligent client. The server only tracks state.**

## Rules

### The Server Should:
- ✅ Validate parameters (types, required fields)
- ✅ Store state in session (unmodified)
- ✅ Track progress (counts, percentages)
- ✅ Return metadata only
- ✅ Handle errors

### The Server Should NEVER:
- ❌ Generate content (ideas, hypotheses, solutions)
- ❌ Analyze content (evaluate, assess, interpret)
- ❌ Make decisions based on content
- ❌ Provide intelligence (infer, deduce, conclude)
- ❌ Transform content (modify, enhance, improve)

## Usage

For MCP servers implementing this pattern, run validation before committing:

```bash
npm run validate:state-tracker
```

## Example Violations

**BAD: Server provides intelligence**
```typescript
// ❌ Generating content
const ideas = this.generateIdeas(prompt);

// ❌ Analyzing content
const quality = this.evaluateQuality(prompt);

// ❌ Making decisions based on content
if (prompt.includes('complex')) {
  return this.useAdvancedMode();
}
```

**GOOD: Server tracks state only**
```typescript
// ✅ Store what LLM provides
const idea = { content: prompt };

// ✅ Use LLM's evaluation (from parameters)
const quality = parameters.quality;

// ✅ Use LLM's decision (from parameters)
const mode = parameters.mode;
```

## Implementation

This pattern ensures:
- Research is reproducible (same state → same behavior)
- Intelligence is auditable (all in LLM prompts/parameters)
- Servers are simple and testable (pure state machines)
- Clear separation of concerns (LLM = intelligence, Server = state)
