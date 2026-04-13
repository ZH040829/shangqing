/**
 * 熵清 V5 - Cloudflare Worker 边缘代理
 * 部署到 Cloudflare Workers (workers.dev 免费)
 * 
 * 使用方式:
 * 1. 在 Cloudflare 创建 API Token (dash.cloudflare.com/profile/api-tokens)
 * 2. 将 token 和 account_id 添加到 GitHub Secrets
 * 3. GitHub Actions 自动部署
 * 4. 访问 https://shangqing.你的account.workers.dev
 */

// 白山智算 EdgeFn API
const EDGEFN_BASE = 'https://api.edgefn.net/v1';
const EDGEFN_KEY = EDGEFN_KEY_HERE; // 从环境变量读取
const MODEL = 'DeepSeek-R1-0528-Qwen3-8B';

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // CORS 预检
    if (request.method === 'OPTIONS') {
      return new Response(null, {
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'POST, GET, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type, Authorization',
          'Access-Control-Max-Age': '86400',
        }
      });
    }

    // 路由
    if (url.pathname === '/v1/chat/completions' && request.method === 'POST') {
      return handleChat(request, env);
    }

    if (url.pathname === '/health') {
      return json({ status: 'ok', service: '熵清 V5 Worker', version: '1.0.0' });
    }

    return json({ error: 'Not found' }, 404);
  }
};

async function handleChat(request, env) {
  try {
    const body = await request.json();
    const messages = body.messages || [];
    const model = body.model || MODEL;

    // 调用白山智算
    const response = await fetch(`${EDGEFN_BASE}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${env.EDGEFN_KEY || EDGEFN_KEY}`,
      },
      body: JSON.stringify({
        model: model,
        messages: messages,
        max_tokens: body.max_tokens || 2000,
        temperature: body.temperature || 0.7,
      })
    });

    const data = await response.json();

    return json(data, response.status, {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'POST, GET, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    });
  } catch (err) {
    return json({ error: err.message || 'Internal error' }, 500, {
      'Access-Control-Allow-Origin': '*',
    });
  }
}

function json(data, status = 200, extraHeaders = {}) {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
      ...extraHeaders
    }
  });
}
