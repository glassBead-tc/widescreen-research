// Express app for health checks and metrics
import express from 'express';
import { pinoHttp, logger } from './utils/logging.js';
import { fetchServiceRegion } from './utils/metadata.js';

const app = express();

// Middleware
app.use(express.json());
app.use(pinoHttp);

// Health check endpoint for Cloud Run
app.get('/health', (req, res) => {
  res.status(200).json({
    status: 'healthy',
    droneType: process.env.DRONE_TYPE || 'generic',
    revision: process.env.K_REVISION || 'unknown',
    timestamp: new Date().toISOString()
  });
});

// Readiness check
app.get('/ready', (req, res) => {
  // Check if MCP server is initialized
  const isReady = global.mcpServerReady || false;
  if (isReady) {
    res.status(200).json({ ready: true });
  } else {
    res.status(503).json({ ready: false, message: 'MCP server not ready' });
  }
});

// Metrics endpoint
app.get('/metrics', async (req, res) => {
  const metrics = {
    droneType: process.env.DRONE_TYPE || 'generic',
    uptime: process.uptime(),
    memory: process.memoryUsage(),
    cpu: process.cpuUsage(),
    region: await fetchServiceRegion() || 'unknown',
    requestsProcessed: global.requestsProcessed || 0
  };
  res.json(metrics);
});

// Root endpoint
app.get('/', (req, res) => {
  const droneInfo = {
    message: 'Widescreen Research MCP Server',
    type: process.env.DRONE_TYPE || 'generic',
    transport: process.env.MCP_TRANSPORT || 'stdio',
    capabilities: getCapabilities()
  };
  res.json(droneInfo);
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

// Error handler
app.use((err, req, res, next) => {
  logger.error('Express error:', err);
  res.status(500).json({ 
    error: 'Internal server error',
    message: process.env.NODE_ENV === 'development' ? err.message : undefined
  });
});

function getCapabilities() {
  const droneType = process.env.DRONE_TYPE || 'generic';
  const capabilities = {
    generic: ['echo', 'ping', 'info'],
    scraper: ['fetch_url', 'extract_data', 'parse_html'],
    processor: ['transform_data', 'validate_data', 'aggregate'],
    research: ['web_search', 'research_papers', 'company_research', 'crawl_url', 
               'find_competitors', 'linkedin_search', 'wikipedia_search', 'github_search'],
    analyzer: ['analyze_text', 'sentiment_analysis', 'extract_entities']
  };
  return capabilities[droneType] || capabilities.generic;
}

export default app; 