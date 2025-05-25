// MCP Server implementation for drone
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { logger } from './utils/logging.js';
import { z } from 'zod';

// Import drone-specific handlers
import { createGenericHandlers } from './drones/generic.js';
import { createScraperHandlers } from './drones/scraper.js';
import { createProcessorHandlers } from './drones/processor.js';
import { createResearchHandlers } from './drones/research.js';

/**
 * Initialize MCP server with appropriate transport
 */
export async function initMcpServer(droneType, transport = 'stdio') {
  try {
    logger.info(`Creating MCP server for drone type: ${droneType}`);
    
    const server = new McpServer({
      name: `drone-mcp-${droneType}`,
      version: '1.0.0'
    });

    logger.info('Getting handlers for drone type');
    const handlers = getHandlersForType(droneType);
    logger.info(`Handlers created: tools=${Object.keys(handlers.tools || {}).length}, resources=${Object.keys(handlers.resources || {}).length}, prompts=${Object.keys(handlers.prompts || {}).length}`);
    
    // Register tools
    if (handlers.tools) {
      for (const [name, handler] of Object.entries(handlers.tools)) {
        server.tool(
          name,
          z.object({}).passthrough(), // Allow any parameters
          async (params) => {
            const result = await handler({ params });
            if (result.error) {
              return {
                content: [{ type: "text", text: `Error: ${result.error.message}` }],
                isError: true
              };
            }
            return {
              content: [{ type: "text", text: JSON.stringify(result.result, null, 2) }]
            };
          }
        );
      }
    }

    // Register resources
    if (handlers.resources) {
      for (const [name, handler] of Object.entries(handlers.resources)) {
        server.resource(
          name,
          `drone://${name}`,
          async (uri) => {
            const result = await handler();
            return result;
          }
        );
      }
    }

    // Register prompts
    if (handlers.prompts) {
      for (const [name, handler] of Object.entries(handlers.prompts)) {
        server.prompt(
          name,
          z.object({}).passthrough(), // Allow any parameters
          async (params) => {
            const result = await handler();
            return result.prompt;
          }
        );
      }
    }

    logger.info('MCP server configured, initializing transport');

    // Initialize transport
    if (transport === 'stdio') {
      const stdioTransport = new StdioServerTransport();
      await server.connect(stdioTransport);
      logger.info('MCP server started with stdio transport');
    } else if (transport === 'http') {
      // Return the server for HTTP middleware
      return server;
    }

    return server;
  } catch (error) {
    logger.error('Error initializing MCP server:', error.message, error.stack);
    throw error;
  }
}

/**
 * Get handlers based on drone type
 */
function getHandlersForType(droneType) {
  try {
    logger.info(`Creating handlers for drone type: ${droneType}`);
    switch (droneType) {
      case 'scraper':
        return createScraperHandlers();
      case 'processor':
        return createProcessorHandlers();
      case 'research':
        return createResearchHandlers();
      case 'generic':
      default:
        return createGenericHandlers();
    }
  } catch (error) {
    logger.error(`Error creating handlers for drone type ${droneType}:`, error.message, error.stack);
    throw error;
  }
}

/**
 * Create Express middleware for HTTP transport
 */
function createHttpMiddleware(server) {
  return async (req, res) => {
    try {
      // Handle MCP protocol over HTTP
      if (req.method === 'POST') {
        const request = req.body;
        const response = await server.handleRequest(request);
        res.json(response);
      } else {
        res.status(405).json({ error: 'Method not allowed' });
      }
    } catch (error) {
      logger.error('HTTP MCP request error:', error);
      res.status(500).json({ error: 'Internal server error' });
    }
  };
} 