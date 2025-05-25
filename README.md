# Widescreen Research - Comprehensive Research MCP Server

A powerful Model Context Protocol (MCP) server that provides comprehensive research capabilities powered by Exa AI. Designed for researchers, analysts, and AI assistants who need access to real-time web search, academic papers, company intelligence, and specialized research tools.

## 🚀 Features

- **Real-time Web Search**: Powered by Exa AI's advanced search capabilities with content extraction
- **Academic Research**: Search and access academic papers and research content
- **Company Intelligence**: Comprehensive company research and competitor analysis
- **Multi-platform Search**: LinkedIn, Wikipedia, GitHub, and general web search
- **Content Extraction**: Direct URL crawling and content analysis
- **Cloud-Ready**: Optimized for Google Cloud Run deployment with auto-scaling
- **MCP Protocol**: Full Model Context Protocol compliance for seamless AI integration

## 🛠️ Research Capabilities

### Core Research Tools

- **`web_search`**: Real-time web search with intelligent content extraction
- **`research_papers`**: Academic paper and research content discovery
- **`company_research`**: Comprehensive company information gathering
- **`crawl_url`**: Extract and analyze content from specific URLs
- **`find_competitors`**: Identify and analyze business competitors
- **`linkedin_search`**: Professional network and company research
- **`wikipedia_search`**: Authoritative encyclopedia content
- **`github_search`**: Open source project and developer research

### Research Specializations

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Academic      │    │   Business      │    │   Technical     │
│   Research      │    │   Intelligence  │    │   Research      │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ • Papers        │    │ • Companies     │    │ • GitHub        │
│ • Citations     │    │ • Competitors   │    │ • Documentation │
│ • Authors       │    │ • Market Data   │    │ • Code Examples │
│ • Institutions  │    │ • LinkedIn      │    │ • Tech Trends   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📋 Prerequisites

- Node.js 16 or later
- Exa AI API key ([Get one here](https://exa.ai))
- Google Cloud Platform account (for deployment)

## 🔧 Installation & Setup

### Local Development

1. **Clone the repository**:
```bash
git clone https://github.com/your-org/widescreen-research.git
cd widescreen-research
```

2. **Install dependencies**:
```bash
npm install
```

3. **Set up environment variables**:
```bash
export EXA_API_KEY="your-exa-api-key"
export DRONE_TYPE="research"
```

4. **Run the server**:
```bash
npm start
```

### Cloud Deployment

Deploy to Google Cloud Run for production use:

```bash
# Build and deploy
export GOOGLE_CLOUD_PROJECT="your-project-id"
export DRONE_TYPE="research"

npm run build-image
npm run deploy
```

## 🎯 Usage

### MCP Client Integration

#### Claude Desktop

Add to your `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "widescreen-research": {
      "command": "node",
      "args": ["/path/to/widescreen-research/index.js"],
      "env": {
        "EXA_API_KEY": "your-exa-api-key",
        "DRONE_TYPE": "research"
      }
    }
  }
}
```

#### Direct MCP Usage

```bash
# Test with MCP Inspector
npx @modelcontextprotocol/inspector node index.js
```

### Research Examples

#### Academic Research
```javascript
// Search for AI safety papers
{
  "method": "tools/call",
  "params": {
    "name": "research_papers",
    "arguments": {
      "query": "AI safety alignment research",
      "numResults": 10,
      "maxCharacters": 5000
    }
  }
}
```

#### Company Intelligence
```javascript
// Research a company and its competitors
{
  "method": "tools/call", 
  "params": {
    "name": "company_research",
    "arguments": {
      "query": "OpenAI company information funding",
      "numResults": 5
    }
  }
}
```

#### Technical Research
```javascript
// Find GitHub repositories
{
  "method": "tools/call",
  "params": {
    "name": "github_search", 
    "arguments": {
      "query": "machine learning frameworks",
      "numResults": 8
    }
  }
}
```

## 🧪 Testing

### Health Check
```bash
curl http://localhost:8080/health
```

### Research Status
```bash
curl http://localhost:8080/
```

### Tool Testing
```bash
# Test web search
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "web_search", "arguments": {"query": "latest AI research", "numResults": 3}}}' | node index.js

# Test company research  
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "company_research", "arguments": {"query": "Tesla", "numResults": 5}}}' | node index.js
```

## 🏗️ Architecture

### Research Server Components

```
├── index.js                 # Main server entry point
├── app.js                   # Express app with health endpoints
├── mcp-server.js            # MCP protocol implementation
├── drones/
│   ├── research.js          # Research tools (Exa AI integration)
│   ├── scraper.js           # Web scraping capabilities
│   ├── processor.js         # Data processing tools
│   └── generic.js           # Basic utilities
└── utils/
    ├── logging.js           # Structured logging
    └── metadata.js          # GCP metadata utilities
```

### Exa AI Integration

The server integrates directly with Exa AI's powerful search API:

- **Real-time Search**: Live web crawling and content extraction
- **Semantic Understanding**: AI-powered result ranking and relevance
- **Content Processing**: Automatic summarization and key information extraction
- **Multi-modal Results**: Text, links, and metadata in structured format

## 🌟 Key Features

✅ **Exa AI Powered**: Advanced search capabilities with real-time web crawling  
✅ **MCP Compliant**: Full Model Context Protocol support for AI integration  
✅ **Cloud Optimized**: Designed for Google Cloud Run with auto-scaling  
✅ **Research Focused**: 8 specialized research tools for different use cases  
✅ **Production Ready**: Comprehensive logging, health checks, and error handling  
✅ **Lightweight**: ~2MB container images with fast cold starts  

## 🔧 Configuration

### Environment Variables

- `EXA_API_KEY`: Your Exa AI API key (required)
- `DRONE_TYPE`: Server type, set to "research" for full capabilities
- `PORT`: Server port (default: 8080)
- `NODE_ENV`: Environment (development/production)
- `LOG_LEVEL`: Logging level (info/debug/error)

### Research Tool Configuration

Each research tool can be configured with parameters:

- `numResults`: Number of search results (1-50)
- `maxCharacters`: Content extraction limit (500-10000)
- `excludeDomain`: Domains to exclude from results
- `category`: Content category filtering

## 🚧 Roadmap

- **Enhanced Analytics**: Research trend analysis and insights
- **Citation Management**: Academic citation formatting and tracking
- **Multi-language Support**: Research in multiple languages
- **Custom Filters**: Advanced search filtering and categorization
- **Research Workflows**: Automated research pipelines
- **Data Export**: Research results in various formats (PDF, CSV, JSON)

## 📚 API Reference

### Research Tools

#### `web_search(query, numResults?)`
Real-time web search with content extraction.

#### `research_papers(query, maxCharacters?, numResults?)`
Search academic papers and research content.

#### `company_research(query, numResults?)`
Comprehensive company information gathering.

#### `crawl_url(url)`
Extract content from specific URLs.

#### `find_competitors(query, excludeDomain?, numResults?)`
Identify business competitors.

#### `linkedin_search(query, numResults?)`
Search LinkedIn for professional content.

#### `wikipedia_search(query, numResults?)`
Search Wikipedia articles.

#### `github_search(query, numResults?)`
Search GitHub repositories and code.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add research capabilities or improvements
4. Test with various research scenarios
5. Submit a pull request

## 📄 License

Apache 2.0 License - see LICENSE file for details.

---

**Powered by Exa AI • Built for Comprehensive Research • MCP Protocol Compliant**
