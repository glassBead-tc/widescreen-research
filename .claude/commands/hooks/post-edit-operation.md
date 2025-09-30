# Post-Edit Operation Hook

## Description

This hook validates that MCP server operations follow the State Tracker Pattern - servers should only track state, never provide intelligence.

## Trigger

Runs after editing operation files in MCP servers that implement state tracking patterns.

## Instructions

### 1. Run Validation

```bash
# For projects with state tracker validation
npm run validate:state-tracker
```

### 2. Review Output

If violations are found:

#### ❌ Errors (Must Fix)
- Content generation (generateIdeas, createSolution, etc.)
- Content analysis (analyzeQuality, evaluatePrompt, etc.)
- Intelligence provision (infer, deduce, conclude)

#### ⚠️ Warnings (Review Carefully)
- Decision making based on content
- Pattern matching on prompt content
- Potential content transformation

### 3. Fix Violations

**BAD: Server tries to be intelligent**
```typescript
// ❌ Generating content
const ideas = this.generateIdeas(prompt);

// ❌ Analyzing content
const quality = this.evaluateQuality(prompt);

// ❌ Making decisions
if (prompt.includes('complex')) {
  return this.useAdvancedMode();
}
```

**GOOD: Server tracks state only**
```typescript
// ✅ Store what LLM provides
const idea = { content: prompt };

// ✅ Use LLM's evaluation
const quality = parameters.quality;

// ✅ Use LLM's decision
const mode = parameters.mode;
```

### 4. Commit Changes

Once validation passes:

```bash
git add src/tools/operations/...
git commit -m "description"
```

The git pre-commit hook will also run validation automatically.

## Checklist

Before committing your changes:

- [ ] Validation passes (`npm run validate:state-tracker`)
- [ ] Server only validates parameters
- [ ] Server only stores state
- [ ] Server only returns progress metadata
- [ ] No content generation
- [ ] No content analysis
- [ ] No intelligent decisions
- [ ] LLM provides all intelligence via prompt/parameters

## Quick Reference

### What the Server Can Do
✅ Validate parameter types and requirements
✅ Store what the LLM sends (unmodified)
✅ Count/aggregate stored items
✅ Calculate progress percentages
✅ Group items by LLM-provided categories
✅ Find items by LLM-provided scores/ratings
✅ Format metadata for display

### What the Server Cannot Do
❌ Generate new content
❌ Analyze content semantics
❌ Make content-based decisions
❌ Evaluate quality/correctness
❌ Infer/deduce/conclude
❌ Transform/enhance content
❌ Suggest improvements
❌ Choose strategies based on content

## Documentation

For more details on the State Tracker Pattern:
- Principle: LLM provides ALL intelligence via prompt/parameters
- Server: Only validates, stores state, tracks progress, returns metadata
- Never: Generate content, analyze semantics, make decisions based on content
