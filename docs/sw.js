const CACHE = 'kunta-mobile-full-20260915-1';
const ASSETS = [
  './','./index.html','./styles.css?v=20260915-full1','./app.js?v=20260915-full1','./manifest.webmanifest','./share.html',
  './assets/kunta-192.png','./assets/kunta-512.png','./assets/kunta-maskable-512.png','./assets/kunta-avatar.png','./assets/botolo.png'
];
self.addEventListener('install', event => {
  event.waitUntil(caches.open(CACHE).then(cache => cache.addAll(ASSETS)));
  self.skipWaiting();
});
self.addEventListener('activate', event => {
  event.waitUntil(caches.keys().then(keys => Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k)))));
  self.clients.claim();
});
self.addEventListener('fetch', event => {
  if (event.request.method !== 'GET') return;
  const req = event.request;
  const isCore = req.mode === 'navigate' || /(?:index\.html|app\.js|styles\.css)$/.test(new URL(req.url).pathname);
  if (isCore) {
    event.respondWith(fetch(req).then(resp => {
      const copy = resp.clone(); caches.open(CACHE).then(c => c.put(req, copy)); return resp;
    }).catch(() => caches.match(req).then(hit => hit || caches.match('./index.html'))));
    return;
  }
  event.respondWith(caches.match(req).then(hit => hit || fetch(req).then(resp => {
    const copy = resp.clone(); caches.open(CACHE).then(c => c.put(req, copy)); return resp;
  })));
});
