const CACHE = 'kunta-mobile-v1';
const ASSETS = [
  './','./index.html','./styles.css','./app.js','./manifest.webmanifest','./share.html',
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
  event.respondWith(caches.match(event.request).then(hit => hit || fetch(event.request).then(resp => {
    const copy = resp.clone();
    caches.open(CACHE).then(c => c.put(event.request, copy));
    return resp;
  }).catch(() => caches.match('./index.html'))));
});
