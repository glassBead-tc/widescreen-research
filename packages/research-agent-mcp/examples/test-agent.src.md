# Test Agent Notebook

<!-- srcbook:{"language":"typescript"} -->

## Overview

Simple test notebook to verify MCP server functionality.

###### hello.ts

```typescript
console.log('Hello from test agent!');
console.log('Environment:', JSON.stringify(process.env, null, 2));
```

###### calculate.ts

```typescript
const a = parseInt(process.env.A || '0');
const b = parseInt(process.env.B || '0');
const sum = a + b;

console.log(`${a} + ${b} = ${sum}`);
console.log(JSON.stringify({ a, b, sum }));
```

###### error-test.ts

```typescript
// This cell intentionally throws an error for testing
throw new Error('Test error - this is expected!');
```
