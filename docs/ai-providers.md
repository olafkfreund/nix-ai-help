# AI Providers Guide

nixai supports **7 AI providers** to give you maximum flexibility in choosing the right AI service for your needs. This guide covers setup, configuration, and best practices for each provider.

## 🎯 Provider Overview

| Provider | Type | Privacy | Speed | Accuracy | Cost | Setup Difficulty |
|----------|------|---------|-------|----------|------|------------------|
| **Ollama** | Local | 🔒 High | ⚡ Fast | ⭐⭐⭐ | 💚 Free | 🔧 Medium |
| **LlamaCpp** | Local | 🔒 High | ⚡ Variable | ⭐⭐⭐ | 💚 Free | 🔧 Hard |
| **Groq** | Cloud | ❌ Low | ⚡⚡⚡ Ultra-fast | ⭐⭐⭐⭐ | 💰 Low-cost | ✅ Easy |
| **Gemini** | Cloud | ❌ Low | ⚡⚡ Fast | ⭐⭐⭐⭐ | 💰 Standard | ✅ Easy |
| **Claude** | Cloud | ❌ Low | ⚡⚡ Fast | ⭐⭐⭐⭐⭐ | 💰💰 Premium | ✅ Easy |
| **OpenAI** | Cloud | ❌ Low | ⚡⚡ Fast | ⭐⭐⭐⭐⭐ | 💰💰 Premium | ✅ Easy |
| **Custom** | Variable | Variable | Variable | Variable | Variable | 🔧 Hard |

## 🔒 Local Providers (Privacy-First)

### Ollama
**Best for**: Privacy-conscious users, offline usage, experimentation

```yaml
# Configuration
ai_provider: ollama
ai_model: llama3

# Available models
models:
  - llama3 (8B, 70B)
  - codellama
  - mistral
  - phi3
  - gemma
```

**Setup:**
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull a model
ollama pull llama3

# Start service
ollama serve
```

**Pros:**
- ✅ Complete privacy (no data leaves your machine)
- ✅ No API costs
- ✅ Works offline
- ✅ Many available models

**Cons:**
- ❌ Requires significant local resources
- ❌ Setup complexity
- ❌ Variable accuracy depending on model

### LlamaCpp
**Best for**: CPU-optimized inference, custom models, resource-constrained environments

```yaml
# Configuration
ai_provider: llamacpp
ai_model: llama-2-7b-chat

# Environment variable
export LLAMACPP_ENDPOINT="http://localhost:8080/completion"
```

**Setup:**
```bash
# Build llama.cpp
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp && make

# Download a model (GGUF format)
wget https://huggingface.co/TheBloke/Llama-2-7b-Chat-GGUF/resolve/main/llama-2-7b-chat.q4_0.gguf

# Start server
./server -m llama-2-7b-chat.q4_0.gguf --port 8080
```

**Pros:**
- ✅ CPU-optimized
- ✅ Lower memory usage than GPU inference
- ✅ Custom model support
- ✅ No API costs

**Cons:**
- ❌ Complex setup
- ❌ Slower than GPU inference
- ❌ Limited to specific model formats

## ⚡ Cloud Providers (Performance & Features)

### Groq
**Best for**: Ultra-fast inference, cost-effective cloud AI, rapid iteration

```yaml
# Configuration
ai_provider: groq
ai_model: llama-3.3-70b-versatile

# Environment variable
export GROQ_API_KEY="your-groq-api-key"
```

**Available Models:**
- `llama-3.3-70b-versatile` (Default) - Latest Llama 3.3 with versatile capabilities
- `llama3-8b-8192` - Fast 8B model with 8K context
- `mixtral-8x7b-32768` - Mixture of experts with 32K context
- `gemma-7b-it` - Google's Gemma instruction-tuned model

**Pros:**
- ✅ Extremely fast inference (fastest available)
- ✅ Cost-effective pricing
- ✅ Good accuracy for most tasks
- ✅ Large context windows

**Cons:**
- ❌ Requires API key
- ❌ Data sent to cloud
- ❌ Limited model selection compared to others

### Google Gemini
**Best for**: Multimodal tasks, strong reasoning, balanced performance

```yaml
# Configuration
ai_provider: gemini
ai_model: gemini-2.5-flash-preview-05-20

# Environment variable
export GEMINI_API_KEY="your-gemini-api-key"
```

**Available Models:**
- `gemini-2.5-flash-preview-05-20` (Default) - Latest optimized model
- `gemini-2.5-pro` - Most capable model for complex tasks
- `gemini-1.5-flash` - Fast model for general tasks

**Pros:**
- ✅ Excellent reasoning capabilities
- ✅ Multimodal support (text, images)
- ✅ Large context windows (1M+ tokens)
- ✅ Good value for performance

**Cons:**
- ❌ Requires API key
- ❌ Google data handling policies
- ❌ Occasional inconsistencies

### Claude (Anthropic)
**Best for**: Complex reasoning, analysis, constitutional AI, detailed explanations

```yaml
# Configuration
ai_provider: claude
ai_model: claude-sonnet-4-20250514

# Environment variable
export CLAUDE_API_KEY="your-claude-api-key"
```

**Available Models:**
- `claude-sonnet-4-20250514` (Default) - Latest and most capable Claude model
- `claude-3-7-sonnet-20250219` - Advanced reasoning with constitutional AI
- `claude-3-5-haiku-20241022` - Fast model for simpler tasks

**Pros:**
- ✅ Excellent for complex analysis and reasoning
- ✅ Constitutional AI approach (helpful, harmless, honest)
- ✅ Superior performance on coding and technical tasks
- ✅ Large context windows (200K tokens)

**Cons:**
- ❌ Higher cost than other providers
- ❌ Requires API key
- ❌ Strict content policies

### OpenAI
**Best for**: Industry-leading performance, complex tasks, production usage

```yaml
# Configuration
ai_provider: openai
ai_model: gpt-4

# Environment variable
export OPENAI_API_KEY="your-openai-api-key"
```

**Available Models:**
- `gpt-4` (Default) - Most capable model for complex reasoning
- `gpt-4-turbo` - Latest GPT-4 with improved performance
- `gpt-3.5-turbo` - Fast and cost-effective for simpler tasks

**Pros:**
- ✅ Industry-leading accuracy
- ✅ Excellent for NixOS-specific questions
- ✅ Strong coding and technical capabilities
- ✅ Large ecosystem and community

**Cons:**
- ❌ Higher cost, especially for GPT-4
- ❌ OpenAI data handling policies
- ❌ Rate limiting on free tier

### Custom Provider
**Best for**: Specialized endpoints, enterprise deployments, experimental models

```yaml
# Configuration
ai_provider: custom
ai_model: custom-model

# Custom configuration
custom_ai:
  base_url: "http://your-api-endpoint:8080/api/generate"
  headers:
    Authorization: "Bearer your-token"
```

**Use Cases:**
- Self-hosted LLM APIs
- Enterprise AI deployments
- Experimental model endpoints
- Custom fine-tuned models

## 🎯 Choosing the Right Provider

### For Privacy-Conscious Users
**Recommended**: Ollama → LlamaCpp
- Keep all data local
- No API costs
- Complete control over models

### For Cost-Effective Performance
**Recommended**: Groq → Gemini → OpenAI
- Start with Groq for ultra-fast, cost-effective inference
- Upgrade to Gemini for better reasoning
- Use OpenAI for most complex tasks

### For Maximum Accuracy
**Recommended**: Claude → OpenAI → Gemini → Groq
- Claude and OpenAI lead in technical accuracy
- Gemini provides good balance of speed and quality
- Groq excels in speed with acceptable accuracy

### For NixOS-Specific Tasks
**Recommended**: 
1. **Complex configurations**: Claude or OpenAI GPT-4
2. **Quick questions**: Groq or Gemini
3. **Learning/experimentation**: Ollama
4. **Privacy-required scenarios**: Ollama or LlamaCpp

## 🔧 Provider Setup Examples

### Quick Setup for Cloud Providers

```bash
# Claude
export CLAUDE_API_KEY="sk-ant-api03-..."
echo 'ai_provider: claude' >> ~/.config/nixai/config.yaml

# Groq
export GROQ_API_KEY="gsk_..."
echo 'ai_provider: groq' >> ~/.config/nixai/config.yaml

# Gemini
export GEMINI_API_KEY="AIza..."
echo 'ai_provider: gemini' >> ~/.config/nixai/config.yaml

# OpenAI
export OPENAI_API_KEY="sk-proj-..."
echo 'ai_provider: openai' >> ~/.config/nixai/config.yaml
```

### Local Provider Setup

```bash
# Ollama (easiest local option)
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull llama3
echo 'ai_provider: ollama' >> ~/.config/nixai/config.yaml

# LlamaCpp (for custom models)
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp && make
./server -m your-model.gguf --port 8080 &
export LLAMACPP_ENDPOINT="http://localhost:8080/completion"
echo 'ai_provider: llamacpp' >> ~/.config/nixai/config.yaml
```

## 🔄 Provider Fallback System

nixai automatically handles provider failures with intelligent fallbacks:

```yaml
# Example fallback configuration
selection_preferences:
  default_provider: "claude"
  task_models:
    nixos_config:
      primary: ["claude:claude-sonnet-4-20250514", "openai:gpt-4"]
      fallback: ["groq:llama-3.3-70b-versatile", "ollama:llama3"]
    general_help:
      primary: ["groq:llama3-8b-8192", "gemini:gemini-2.5-flash-preview-05-20"]
      fallback: ["ollama:llama3"]
```

**Fallback Logic:**
1. Try primary provider/model
2. If API key missing or service unavailable, try fallback
3. Always fall back to Ollama as final option (if available)

## 📊 Provider Performance Comparison

### Speed Comparison (typical response times)
- **Groq**: 0.5-2 seconds ⚡⚡⚡
- **Gemini**: 1-3 seconds ⚡⚡
- **Claude**: 2-5 seconds ⚡⚡
- **OpenAI**: 2-6 seconds ⚡⚡
- **Ollama**: 3-15 seconds ⚡ (depends on hardware)
- **LlamaCpp**: 5-30 seconds ⚡ (depends on hardware)

### Context Window Comparison
- **Gemini**: Up to 1M tokens
- **Claude**: Up to 200K tokens
- **OpenAI**: Up to 128K tokens
- **Groq**: Up to 32K tokens
- **Ollama**: Up to 128K tokens (model dependent)
- **LlamaCpp**: Up to 128K tokens (model dependent)

### Cost Comparison (estimated per 1M tokens)
- **Ollama**: $0 (hardware costs only)
- **LlamaCpp**: $0 (hardware costs only)
- **Groq**: ~$0.27-$0.59
- **Gemini**: ~$1.25-$2.50
- **Claude**: ~$3.00-$15.00
- **OpenAI**: ~$0.50-$30.00

## 🛠️ Troubleshooting

### Common Issues

**API Key Not Found:**
```bash
# Check environment variables
env | grep -E "(CLAUDE|GROQ|GEMINI|OPENAI)_API_KEY"

# Set permanently
echo 'export CLAUDE_API_KEY="your-key"' >> ~/.bashrc
source ~/.bashrc
```

**Local Provider Connection Failed:**
```bash
# Check Ollama status
ollama list
curl http://localhost:11434/api/tags

# Check LlamaCpp server
curl http://localhost:8080/health
```

**Provider Timeout:**
```yaml
# Increase timeout in configuration
ai_timeouts:
  claude: 60    # Increase from default 30s
  groq: 45      # Increase from default 30s
```

**Model Not Found:**
```bash
# List available models
ollama list  # For Ollama
# Check API documentation for cloud providers
```

## 📚 Additional Resources

- [Configuration Guide](config.md)
- [Environment Variables](../README.md#environment-variables)
- [Provider API Documentation](#provider-apis)
- [Troubleshooting Guide](../TROUBLESHOOTING.md)

---

*This guide covers nixai v1.0.4 with support for 7 AI providers. For the latest updates, see the [changelog](../configs/changelog.yaml).*
