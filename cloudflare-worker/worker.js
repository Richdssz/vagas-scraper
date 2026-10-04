// Cloudflare Worker: marca uma vaga como "já vi" gravando em vagas_marcadas.json no GitHub.
// Variáveis (wrangler secret put): GITHUB_TOKEN, MARCAR_SECRET
// Variáveis (wrangler.toml [vars]): GITHUB_REPO (ex: "Richdssz/vagas-scraper"), GITHUB_BRANCH (ex: "main")

const ARQUIVO = "vagas_marcadas.json";

async function hmacHex(secret, msg) {
  const key = await crypto.subtle.importKey(
    "raw", new TextEncoder().encode(secret),
    { name: "HMAC", hash: "SHA-256" }, false, ["sign"]
  );
  const sig = await crypto.subtle.sign("HMAC", key, new TextEncoder().encode(msg));
  return [...new Uint8Array(sig)].map(b => b.toString(16).padStart(2, "0")).join("");
}

function safeEqual(a, b) {
  if (a.length !== b.length) return false;
  let r = 0;
  for (let i = 0; i < a.length; i++) r |= a.charCodeAt(i) ^ b.charCodeAt(i);
  return r === 0;
}

function pagina(titulo, msg, ok) {
  const cor = ok ? "#34d399" : "#f87171";
  const html = `<!DOCTYPE html><html lang="pt-BR"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>${titulo}</title></head>
<body style="margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;background:#09090b;font-family:-apple-system,Segoe UI,Roboto,sans-serif;color:#fafafa">
<div style="max-width:420px;margin:24px;padding:32px;background:#121215;border:1px solid #27272a;border-radius:12px;text-align:center">
<div style="font-size:40px;color:${cor}">${ok ? "&#10003;" : "&#10007;"}</div>
<h1 style="font-size:20px;margin:12px 0 8px">${titulo}</h1>
<p style="margin:0;color:#a1a1aa;font-size:14px;line-height:1.5">${msg}</p></div></body></html>`;
  return new Response(html, { status: ok ? 200 : 400, headers: { "content-type": "text/html; charset=utf-8" } });
}

function esc(s) {
  return String(s).replace(/[&<>"]/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
}

async function gravarNoGitHub(env, id, titulo) {
  const api = `https://api.github.com/repos/${env.GITHUB_REPO}/contents/${ARQUIVO}`;
  const headers = {
    Authorization: `Bearer ${env.GITHUB_TOKEN}`,
    Accept: "application/vnd.github+json",
    "User-Agent": "vagas-scraper-worker",
  };
  const branch = env.GITHUB_BRANCH || "main";

  for (let tentativa = 0; tentativa < 3; tentativa++) {
    let sha, lista = [];
    const get = await fetch(`${api}?ref=${branch}`, { headers });
    if (get.status === 200) {
      const j = await get.json();
      sha = j.sha;
      try { lista = JSON.parse(atob(j.content.replace(/\n/g, "")) || "[]"); } catch { lista = []; }
    } else if (get.status !== 404) {
      throw new Error(`GitHub GET ${get.status}`);
    }

    if (lista.some(x => x.id === id)) return; // idempotente

    lista.push({ id, titulo, marcado_em: new Date().toISOString() });
    const body = {
      message: `chore: marcar vaga como vista [skip ci]`,
      content: btoa(unescape(encodeURIComponent(JSON.stringify(lista, null, 2)))),
      branch,
      ...(sha ? { sha } : {}),
    };
    const put = await fetch(api, { method: "PUT", headers, body: JSON.stringify(body) });
    if (put.ok) return;
    if (put.status !== 409 && put.status !== 422) throw new Error(`GitHub PUT ${put.status}`);
    // conflito de sha: tenta de novo
  }
  throw new Error("Conflito ao gravar no GitHub");
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname !== "/marcar") return new Response("Not found", { status: 404 });

    const id = url.searchParams.get("id") || "";
    const titulo = (url.searchParams.get("t") || "").slice(0, 200);
    const sig = url.searchParams.get("sig") || "";
    if (!id || !sig) return pagina("Link inválido", "Parâmetros ausentes.", false);

    const esperado = await hmacHex(env.MARCAR_SECRET, id);
    if (!safeEqual(esperado, sig)) return pagina("Assinatura inválida", "Este link não é autêntico.", false);

    try {
      await gravarNoGitHub(env, id, titulo);
    } catch (e) {
      return pagina("Falha ao salvar", esc(e.message), false);
    }
    return pagina("Vaga marcada como vista", `<strong>${esc(titulo || id)}</strong><br>Você não verá mais esta vaga nos próximos e-mails.`, true);
  },
};
