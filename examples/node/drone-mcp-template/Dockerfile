# Use the official lightweight Node.js image.
FROM node:22-slim

# Create and change to the app directory.
WORKDIR /usr/src/app

# Copy application dependency manifests to the container image.
COPY package*.json ./

# Install production dependencies.
RUN npm ci --only=production

# Copy local code to the container image.
COPY . ./

# Set up non-root user for security
RUN useradd -r -s /bin/false nodeuser && \
    chown -R nodeuser:nodeuser /usr/src/app
USER nodeuser

# The drone type can be set at runtime
ENV DRONE_TYPE=generic
ENV MCP_TRANSPORT=stdio
ENV NODE_ENV=production

# Run the web service on container startup.
ENTRYPOINT ["node", "index.js"] 