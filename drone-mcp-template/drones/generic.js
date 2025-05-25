// Generic drone MCP handlers
import { logger } from '../utils/logging.js';

export function createGenericHandlers() {
  return {
    tools: {
      echo: async (request) => {
        const { message } = request.params;
        logger.info('Echo tool called', { message });
        return {
          result: {
            echo: message,
            timestamp: new Date().toISOString()
          }
        };
      },

      ping: async (request) => {
        logger.info('Ping tool called');
        return {
          result: {
            status: 'pong',
            droneType: process.env.DRONE_TYPE || 'generic',
            timestamp: new Date().toISOString()
          }
        };
      },

      info: async (request) => {
        logger.info('Info tool called');
        return {
          result: {
            droneType: process.env.DRONE_TYPE || 'generic',
            revision: process.env.K_REVISION || 'local',
            service: process.env.K_SERVICE || 'unknown',
            uptime: process.uptime(),
            memory: process.memoryUsage()
          }
        };
      }
    },

    resources: {
      status: async () => {
        return {
          contents: [{
            uri: 'drone://status',
            mimeType: 'application/json',
            text: JSON.stringify({
              healthy: true,
              droneType: process.env.DRONE_TYPE || 'generic',
              uptime: process.uptime()
            }, null, 2)
          }]
        };
      }
    },

    prompts: {
      help: async () => {
        return {
          prompt: {
            name: 'help',
            description: 'Get help on using this drone',
            arguments: [],
            content: `This is a ${process.env.DRONE_TYPE || 'generic'} drone MCP server.
            
Available tools:
- echo: Echo back a message
- ping: Check if the drone is responsive
- info: Get drone information

Available resources:
- status: Get current drone status`
          }
        };
      }
    }
  };
} 