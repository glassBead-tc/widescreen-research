FROM node:20-alpine

WORKDIR /workspace

# Copy package files
COPY package.json tsconfig.json ./
COPY src/ ./src/

# Install dependencies
RUN npm install

# Build TypeScript
RUN npm run build

# Copy notebook (will be overridden by volume mount)
COPY examples/research-agent-example.src.md ./notebook.src.md

# Runtime environment
ENV AGENT_ID=agent-1
ENV NOTEBOOK_PATH=./notebook.src.md
ENV WORKDIR=/workspace

CMD ["node", "dist/index.js"]