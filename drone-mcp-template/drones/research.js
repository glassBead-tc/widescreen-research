// Research drone MCP handlers - connects to Exa API directly
import { logger } from '../utils/logging.js';
import axios from 'axios';

const EXA_API_BASE_URL = 'https://api.exa.ai';

/**
 * Make a direct call to Exa API
 */
async function callExaAPI(endpoint, params) {
  try {
    const exaApiKey = process.env.EXA_API_KEY;
    if (!exaApiKey) {
      throw new Error('EXA_API_KEY environment variable is required for research drone');
    }

    logger.info(`Calling Exa API: ${endpoint}`, { params });

    const response = await axios.post(`${EXA_API_BASE_URL}${endpoint}`, params, {
      headers: {
        'accept': 'application/json',
        'content-type': 'application/json',
        'x-api-key': exaApiKey
      },
      timeout: 30000
    });

    logger.info(`Received response from Exa API: ${endpoint}`);
    return response.data;
  } catch (error) {
    logger.error(`Error calling Exa API ${endpoint}:`, error);
    throw error;
  }
}

export function createResearchHandlers() {
  // Mark MCP server as ready immediately
  global.mcpServerReady = true;

  return {
    tools: {
      web_search: async (request) => {
        try {
          const { query, numResults = 5 } = request.params;
          
          const result = await callExaAPI('/search', {
            query,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters: 3000
              },
              livecrawl: 'always'
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'WEB_SEARCH_ERROR',
              message: error.message
            }
          };
        }
      },

      research_papers: async (request) => {
        try {
          const { query, maxCharacters = 3000, numResults = 5 } = request.params;
          
          const result = await callExaAPI('/search', {
            query: `${query} academic research paper`,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters
              }
            },
            category: 'research paper'
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'RESEARCH_PAPERS_ERROR',
              message: error.message
            }
          };
        }
      },

      company_research: async (request) => {
        try {
          const { query, numResults = 5 } = request.params;
          
          const result = await callExaAPI('/search', {
            query: `${query} company information`,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters: 5000
              },
              livecrawl: 'always'
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'COMPANY_RESEARCH_ERROR',
              message: error.message
            }
          };
        }
      },

      crawl_url: async (request) => {
        try {
          const { url } = request.params;
          
          const result = await callExaAPI('/contents', {
            ids: [url],
            text: {
              maxCharacters: 10000
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'CRAWL_URL_ERROR',
              message: error.message
            }
          };
        }
      },

      find_competitors: async (request) => {
        try {
          const { query, excludeDomain, numResults = 10 } = request.params;
          
          let searchQuery = `${query} competitors similar companies`;
          if (excludeDomain) {
            searchQuery += ` -site:${excludeDomain}`;
          }
          
          const result = await callExaAPI('/search', {
            query: searchQuery,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters: 2000
              }
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'FIND_COMPETITORS_ERROR',
              message: error.message
            }
          };
        }
      },

      linkedin_search: async (request) => {
        try {
          const { query, numResults = 5 } = request.params;
          
          const result = await callExaAPI('/search', {
            query: `${query} site:linkedin.com`,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters: 2000
              }
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'LINKEDIN_SEARCH_ERROR',
              message: error.message
            }
          };
        }
      },

      wikipedia_search: async (request) => {
        try {
          const { query, numResults = 5 } = request.params;
          
          const result = await callExaAPI('/search', {
            query: `${query} site:wikipedia.org`,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters: 5000
              }
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'WIKIPEDIA_SEARCH_ERROR',
              message: error.message
            }
          };
        }
      },

      github_search: async (request) => {
        try {
          const { query, numResults = 5 } = request.params;
          
          const result = await callExaAPI('/search', {
            query: `${query} site:github.com`,
            type: 'auto',
            numResults,
            contents: {
              text: {
                maxCharacters: 3000
              }
            }
          });

          return { result };
        } catch (error) {
          return {
            error: {
              code: 'GITHUB_SEARCH_ERROR',
              message: error.message
            }
          };
        }
      }
    },

    resources: {
      'research-status': async () => {
        return {
          contents: [{
            uri: 'drone://research-status',
            mimeType: 'application/json',
            text: JSON.stringify({
              connected: true,
              droneType: 'research',
              exaApiKeySet: !!process.env.EXA_API_KEY,
              apiEndpoint: EXA_API_BASE_URL,
              availableTools: [
                'web_search',
                'research_papers', 
                'company_research',
                'crawl_url',
                'find_competitors',
                'linkedin_search',
                'wikipedia_search',
                'github_search'
              ]
            }, null, 2)
          }]
        };
      }
    },

    prompts: {
      'research-help': async () => {
        return {
          prompt: {
            name: 'research-help',
            description: 'Get help on using the research drone',
            arguments: [],
            content: `This is a research drone powered by Exa AI.

Available research tools:
- web_search: Real-time web search with content extraction
- research_papers: Search academic papers and research content
- company_research: Comprehensive company information gathering
- crawl_url: Extract content from specific URLs
- find_competitors: Identify competitors for a company
- linkedin_search: Search LinkedIn for companies and people
- wikipedia_search: Search Wikipedia articles
- github_search: Search GitHub repositories

Each tool leverages Exa's powerful search and content extraction capabilities.`
          }
        };
      }
    }
  };
} 