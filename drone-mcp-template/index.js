// Widescreen Research MCP Server
import { createServer } from 'http';
import app from './app.js';
import { initMcpServer } from './mcp-server.js';
import { logger, initLogCorrelation } from './utils/logging.js';
import { fetchProjectId } from './utils/metadata.js';

/**
 * Initialize app and start both Express and MCP servers
 */
const main = async () => {
  try {
    // Get project ID for logging correlation
    let project = process.env.GOOGLE_CLOUD_PROJECT;
    if (!project) {
      try {
        project = await fetchProjectId();
      } catch (err) {
        logger.warn('Could not fetch Project Id for tracing.', err);
      }
    }
    
    // Initialize request-based logger with project Id
    initLogCorrelation(project);

    // Get configuration from environment
    const PORT = process.env.PORT || 8080;
    const DRONE_TYPE = process.env.DRONE_TYPE || 'generic';
    const COORDINATOR_URL = process.env.COORDINATOR_URL;
    const MCP_TRANSPORT = process.env.MCP_TRANSPORT || 'stdio'; // 'stdio' or 'http'

    logger.info({
      message: 'Starting Widescreen Research MCP server',
      droneType: DRONE_TYPE,
      transport: MCP_TRANSPORT,
      coordinatorUrl: COORDINATOR_URL
    });

    // Initialize MCP server based on transport type
    if (MCP_TRANSPORT === 'http') {
      // For HTTP transport, integrate MCP with Express
      logger.info('Initializing MCP server with HTTP transport');
      const mcpHandler = await initMcpServer(DRONE_TYPE, 'http');
      app.use('/mcp', mcpHandler);
    } else {
      // For stdio transport, run MCP server separately
      logger.info('Initializing MCP server with stdio transport');
      await initMcpServer(DRONE_TYPE, 'stdio');
    }

    // Start HTTP server for health checks and metrics
    const server = createServer(app);
    server.listen(PORT, () => {
      logger.info(`HTTP server listening on port ${PORT}`);
      
      // Report to coordinator if URL is provided
      if (COORDINATOR_URL) {
        reportToCoordinator(COORDINATOR_URL, DRONE_TYPE, PORT);
      }
    });
  } catch (error) {
    logger.error('Failed to start server:', error);
    process.exit(1);
  }
};

/**
 * Report drone status to coordinator
 */
async function reportToCoordinator(coordinatorUrl, droneType, port) {
  try {
    const { authenticatedRequest } = await import('./utils/metadata.js');
    const response = await authenticatedRequest(
      `${coordinatorUrl}/api/drones/register`,
      'POST',
      {
        droneType,
        serviceUrl: process.env.K_SERVICE ? `https://${process.env.K_SERVICE}` : `http://localhost:${port}`,
        revision: process.env.K_REVISION || 'local',
        capabilities: getDroneCapabilities(droneType)
      }
    );
    logger.info('Successfully registered with coordinator', response.data);
  } catch (error) {
    logger.error('Failed to register with coordinator', error);
  }
}

/**
 * Get drone capabilities based on type
 */
function getDroneCapabilities(droneType) {
  const capabilities = {
    generic: ['echo', 'ping'],
    scraper: ['fetch_url', 'extract_data'],
    processor: ['transform_data', 'validate_data'],
    research: ['web_search', 'research_papers', 'company_research', 'crawl_url', 
               'find_competitors', 'linkedin_search', 'wikipedia_search', 'github_search'],
    analyzer: ['analyze_text', 'sentiment_analysis']
  };
  return capabilities[droneType] || capabilities.generic;
}

/**
 * Listen for termination signal
 */
process.on('SIGTERM', () => {
  logger.info('Caught SIGTERM, shutting down gracefully');
  logger.flush();
  process.exit(0);
});

// Handle uncaught exceptions
process.on('uncaughtException', (err) => {
  logger.error('Uncaught exception:', err);
  logger.flush();
  process.exit(1);
});

main().catch(err => {
  logger.error('Failed to start server:', err);
  process.exit(1);
}); 