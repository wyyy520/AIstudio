const http = require('http');
const https = require('https');

const DASHSCOPE_KEY = 'ms-a2e526f5-cc8b-4ffe-bf45-fb95d3373836';
const DASHSCOPE_API = 'dashscope.aliyuncs.com';
const PORT = 4000;

function anthropicToOpenAI(body) {
  const messages = [];
  if (body.system) {
    messages.push({ role: 'system', content: body.system });
  }
  for (const msg of body.messages || []) {
    if (msg.role === 'assistant' && msg.content === '') continue;
    messages.push({
      role: msg.role === 'user' ? 'user' : 'assistant',
      content: typeof msg.content === 'string' ? msg.content : (msg.content || []).map(c => c.text || '').join(''),
    });
  }
  return {
    model: 'qwen3-32b',
    messages,
    max_tokens: body.max_tokens || 4096,
    temperature: body.temperature ?? 0.7,
    stream: body.stream || false,
  };
}

function openAIToAnthropic(openAIResp) {
  const choice = openAIResp.choices?.[0];
  if (!choice) return { content: [{ type: 'text', text: '' }], stop_reason: 'end_turn' };
  return {
    id: openAIResp.id || 'msg_proxy',
    type: 'message',
    role: 'assistant',
    content: [{ type: 'text', text: choice.message?.content || '' }],
    model: 'qwen3-32b',
    stop_reason: choice.finish_reason === 'stop' ? 'end_turn' : choice.finish_reason || 'end_turn',
    stop_sequence: null,
    usage: openAIResp.usage || {},
  };
}

function postToDashScope(path, body) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify(body);
    const options = {
      hostname: DASHSCOPE_API,
      path,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + DASHSCOPE_KEY,
        'Content-Length': Buffer.byteLength(data),
      },
    };
    const req = https.request(options, (res) => {
      let chunks = [];
      res.on('data', (c) => chunks.push(c));
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, body: JSON.parse(Buffer.concat(chunks).toString()) });
        } catch (e) {
          resolve({ status: res.statusCode, body: Buffer.concat(chunks).toString() });
        }
      });
    });
    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

const server = http.createServer(async (req, res) => {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Headers', '*');
  if (req.method === 'OPTIONS') { res.writeHead(204); res.end(); return; }

  if (req.method === 'GET' && req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ status: 'ok' }));
    return;
  }

  if (req.method === 'POST' && req.url === '/v1/messages') {
    let body = '';
    req.on('data', (c) => body += c);
    req.on('end', async () => {
      try {
        const anthropicReq = JSON.parse(body);
        const openAIReq = anthropicToOpenAI(anthropicReq);
        const dashScopeResp = await postToDashScope('/compatible-mode/v1/chat/completions', openAIReq);
        
        if (dashScopeResp.status !== 200) {
          console.error('DashScope error:', dashScopeResp.status, JSON.stringify(dashScopeResp.body));
          res.writeHead(dashScopeResp.status, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({ error: { message: 'DashScope API error', status: dashScopeResp.status, details: dashScopeResp.body } }));
          return;
        }
        
        const anthropicResp = openAIToAnthropic(dashScopeResp.body);
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(anthropicResp));
      } catch (e) {
        console.error('Proxy error:', e);
        res.writeHead(500, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: { message: e.message } }));
      }
    });
    return;
  }

  res.writeHead(404); res.end();
});

server.listen(PORT, '127.0.0.1', () => {
  console.log('Anthropic -> DashScope proxy running on http://127.0.0.1:' + PORT);
  console.log('Configure Claude Code with:');
  console.log(JSON.stringify({ model: 'qwen3-32b', baseURL: 'http://127.0.0.1:' + PORT, apiKey: 'any' }, null, 2));
});
