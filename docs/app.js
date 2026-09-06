'use strict';

const OPS = [
  ['Lettere indicate', 'Scrivi le lettere da contare, es. aeiou'],
  ['Caratteri', 'Conta caratteri Unicode e byte UTF-8'],
  ['Parole', 'Conta parole e parole distinte'],
  ['Righe', 'Conta righe e righe non vuote'],
  ['Spazi', 'Conta spazi semplici e caratteri bianchi'],
  ['Cifre', 'Conta le cifre'],
  ['Punteggiatura', 'Conta i segni di punteggiatura'],
  ['Occorrenze parola/stringa', 'Parametro: parola o frase da cercare'],
  ['Parole uniche', 'Elenca le parole distinte'],
  ['Parole ripetute', 'Elenca solo le parole che ricorrono'],
  ['Frequenza lettere', 'Classifica le lettere per frequenza'],
  ['Frequenza parole', 'Classifica le parole per frequenza'],
  ['Parola più corta / più lunga', 'Trova gli estremi di lunghezza'],
  ['Palindromi', 'Parametro: lunghezza minima, default 3'],
  ['Bifronti', 'Parametro: lunghezza minima, default 3'],
  ['Anagrammi', 'Parametro: lunghezza minima, default 3'],
  ['Acrostico', 'Prima lettera/cifra utile di ogni riga non vuota'],
  ['Telestico', 'Ultima lettera/cifra utile di ogni riga non vuota'],
  ['Inverti testo', 'Inverte tutti i caratteri'],
  ['Inverti ordine parole', 'Inverte la sequenza delle parole'],
  ['Ordina parole alfabeticamente', 'Visualizza subito ripetizioni e famiglie lessicali'],
  ['Estrai numeri', 'Estrae le sequenze numeriche'],
  ['Schema rime / desinenze', 'Parametro: profondità della coda, default 3'],
  ['Rima baciata / inclusione progressiva', 'Parametro: nucleo minimo, default 3'],
  ['LetterTransport', 'Parametro: trasporto minimo, default 3'],
  ['Sciarade / salti di dominio', 'Parametro: lunghezza minima segmento, default 2'],
  ['Kunta il testo (fenomeni)', 'Ricognizione combinata di strutture e ricorrenze']
];

const $ = id => document.getElementById(id);
const input = $('inputText'), output = $('outputText'), opSel = $('operation'), param = $('param');

OPS.forEach(([name], i) => {
  const o = document.createElement('option'); o.value = i; o.textContent = name; opSel.appendChild(o);
});
opSel.value = '20';

function updateHint(){ $('hint').textContent = OPS[Number(opSel.value)][1]; }
opSel.addEventListener('change', updateHint); updateHint();

function normalizeCase(s, sensitive){ return sensitive ? s : s.toLowerCase(); }
function normalizeLines(s){ return s.replace(/\r\n?/g,'\n').split('\n'); }
function tokenizeWords(s){ return s.match(/[\p{L}\p{N}]+(?:[’'\-][\p{L}\p{N}]+)*/gu) || []; }
function cleanWord(s, sensitive){
  let w = s.replace(/^[^\p{L}\p{N}]+|[^\p{L}\p{N}]+$/gu,'');
  return normalizeCase(w, sensitive);
}
function reverseRunes(s){ return Array.from(s).reverse().join(''); }
function parsePositive(s, def, min, max){ const n = Number.parseInt(String(s).trim(),10); return Number.isFinite(n) ? Math.max(min,Math.min(max,n)) : def; }
function parseMinLen(s){ return parsePositive(s,3,1,100); }
function limitLines(s,max){ const a=s.split('\n'); return a.length<=max?s:a.slice(0,max).join('\n')+`\n\n… output limitato a ${max} righe.`; }
function freqWords(words, sensitive){
  const freq=new Map(), display=new Map();
  for(const w0 of words){ const k=cleanWord(w0,sensitive); if(!k) continue; freq.set(k,(freq.get(k)||0)+1); if(!display.has(k)) display.set(k,w0); }
  return {freq,display};
}
function distinctWordCount(words,sensitive){ return freqWords(words,sensitive).freq.size; }
function sortedEntries(map, cmp){ return [...map.entries()].sort(cmp); }
function countOccurrences(hay,needle){ if(!needle) return 0; let n=0,pos=0; while((pos=hay.indexOf(needle,pos))!==-1){n++;pos+=needle.length;} return n; }

function findPalindromes(words,sensitive,minLen){
  const {freq,display}=freqWords(words,sensitive); const rows=[]; let total=0;
  for(const [k,n] of freq){ if(Array.from(k).length>=minLen && k===reverseRunes(k)){rows.push([k,n]);total+=n;} }
  rows.sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0]));
  return `Palindromi distinti: ${rows.length}\nOccorrenze totali: ${total}\nLunghezza minima: ${minLen}\n\n`+rows.map(([k,n])=>`${display.get(k)} = ${n}`).join('\n');
}
function findBifronti(words,sensitive,minLen){
  const {freq,display}=freqWords(words,sensitive); const seen=new Set(), pairs=[];
  for(const k of freq.keys()){
    if(Array.from(k).length<minLen) continue; const rev=reverseRunes(k); if(rev===k||!freq.has(rev)) continue;
    const key=[k,rev].sort().join('\0'); if(seen.has(key)) continue; seen.add(key); pairs.push([display.get(k),display.get(rev)]);
  }
  pairs.sort((a,b)=>a[0].localeCompare(b[0],undefined,{sensitivity:'base'}));
  return `Bifronti trovati: ${pairs.length}\nLunghezza minima: ${minLen}\n\n`+pairs.map(p=>`${p[0]} ↔ ${p[1]}`).join('\n');
}
function sortedRunes(s){ return Array.from(s).sort().join(''); }
function findAnagrams(words,sensitive,minLen){
  const {display}=freqWords(words,sensitive); const groups=new Map();
  for(const [k,disp] of display){ if(Array.from(k).length<minLen) continue; const sig=sortedRunes(k); if(!groups.has(sig)) groups.set(sig,[]); groups.get(sig).push(disp); }
  const arr=[...groups.values()].filter(a=>a.length>1).map(a=>a.sort((x,y)=>x.localeCompare(y,undefined,{sensitivity:'base'}))).sort((a,b)=>a[0].localeCompare(b[0],undefined,{sensitivity:'base'}));
  return limitLines(`Gruppi di anagrammi: ${arr.length}\nLunghezza minima: ${minLen}\n\n`+arr.map(a=>a.join(' · ')).join('\n'),700);
}
function runeSuffix(s,n){ const r=Array.from(s); return r.length<=n?s:r.slice(-n).join(''); }
function finalWordOfLine(line,sensitive){ const w=tokenizeWords(line); return w.length?cleanWord(w[w.length-1],sensitive):''; }
function rhymeLabel(n){ const labels='xyzwvutsrqponmlkjihgfedcba'; return n<labels.length?labels[n]:`r${n+1}`; }
function analyzeRhymeScheme(text,sensitive,depth){
  const lines=normalizeLines(text), stanzas=[]; let cur=[];
  const flush=()=>{if(cur.length){stanzas.push(cur);cur=[];}};
  for(const line of lines){ if(!line.trim()) flush(); else cur.push(line); } flush();
  if(!stanzas.length) return 'Nessun verso rilevato.';
  const out=[`Rima = desinenza grafica · profondità: ${depth} lettere`,`Strofe rilevate: ${stanzas.length}`,''];
  stanzas.forEach((stanza,si)=>{
    const rows=[], counts=new Map(), order=[];
    for(const line of stanza){ const word=finalWordOfLine(line,sensitive); if(!word) continue; const suffix=runeSuffix(word,depth); rows.push({word,suffix}); counts.set(suffix,(counts.get(suffix)||0)+1); if(!order.includes(suffix)) order.push(suffix); }
    if(!rows.length) return;
    const groups=new Map();
    if(order.length===2){ let singleton='',dominant=''; for(const s of order){ if(counts.get(s)===1) singleton=s; else dominant=s; } if(singleton&&dominant){groups.set(dominant,'x');groups.set(singleton,'z');} }
    if(!groups.size) order.forEach((s,i)=>groups.set(s,rhymeLabel(i)));
    rows.forEach(r=>r.label=groups.get(r.suffix));
    if(stanzas.length>1) out.push(`STROFA ${si+1}`);
    rows.forEach((r,i)=>out.push(`${String(i+1).padStart(2)}. ${r.word.padEnd(22)} → -${r.suffix.padEnd(10)} → ${r.label}`));
    out.push('',`Schema: ${rows.map(r=>r.label).join('-')}`,'Coda progressiva:');
    for(let d=1;d<=depth;d++){
      const f=new Map(); rows.forEach(r=>{const s=runeSuffix(r.word,d);f.set(s,(f.get(s)||0)+1)});
      const parts=[...f.entries()].filter(([,n])=>n>1).sort((a,b)=>a[0].localeCompare(b[0])).map(([s,n])=>`-${s} ×${n}`);
      if(parts.length) out.push(`  ${d}: ${parts.join(' · ')}`);
    }
    if(si<stanzas.length-1) out.push('');
  });
  return limitLines(out.join('\n'),700);
}
function findInclusionChains(words,sensitive,minLen){
  const {display}=freqWords(words,sensitive); const keys=[...display.keys()].filter(k=>Array.from(k).length>=minLen).sort((a,b)=>Array.from(a).length-Array.from(b).length||a.localeCompare(b));
  const children=new Map(), incoming=new Map();
  for(const a of keys){
    const la=Array.from(a).length; const candidates=keys.filter(b=>Array.from(b).length>la&&b.endsWith(a));
    for(const b of candidates){ const lb=Array.from(b).length; let immediate=true;
      for(const c of candidates){ const lc=Array.from(c).length; if(c!==b&&lc>la&&lc<lb&&c.endsWith(a)&&b.endsWith(c)){immediate=false;break;} }
      if(immediate){ if(!children.has(a)) children.set(a,[]); children.get(a).push(b); incoming.set(b,(incoming.get(b)||0)+1); }
    }
    if(children.has(a)) children.get(a).sort((x,y)=>Array.from(x).length-Array.from(y).length||x.localeCompare(y));
  }
  const chains=[];
  function dfs(node,path){ const p=[...path,node], ch=children.get(node)||[]; if(!ch.length){if(p.length>=2) chains.push(p);return;} ch.forEach(c=>dfs(c,p)); }
  keys.filter(k=>!(incoming.get(k)>0)&&(children.get(k)||[]).length).forEach(k=>dfs(k,[]));
  chains.sort((a,b)=>b.length-a.length||a[0].localeCompare(b[0]));
  const out=[`Rima baciata · inclusione progressiva`,`Nucleo minimo: ${minLen} lettere`,`Catene trovate: ${chains.length}`,''];
  chains.slice(0,100).forEach((c,i)=>{out.push(`${String(i+1).padStart(2)}. ${c.map(k=>display.get(k)).join(' → ')}`,`    nucleo: ${display.get(c[0])} · profondità: ${c.length}`)});
  if(!chains.length) out.push('Nessuna inclusione progressiva rilevata.');
  if(chains.length>100) out.push('\n… altre catene omesse.');
  return out.join('\n');
}
function longestCommonSubstring(a,b){
  const ar=Array.from(a), br=Array.from(b); if(!ar.length||!br.length) return '';
  let prev=new Array(br.length+1).fill(0), bestLen=0,bestEnd=0;
  for(let i=1;i<=ar.length;i++){ const cur=new Array(br.length+1).fill(0); for(let j=1;j<=br.length;j++){ if(ar[i-1]===br[j-1]){cur[j]=prev[j-1]+1;if(cur[j]>bestLen){bestLen=cur[j];bestEnd=i;}} } prev=cur; }
  return ar.slice(bestEnd-bestLen,bestEnd).join('');
}
function findLetterTransport(words,sensitive,minLen){
  if(words.length<2) return 'Servono almeno due parole.'; const out=[`LetterTransport · parole consecutive`,`Trasporto minimo: ${minLen} lettere`,'']; let found=0;
  for(let i=0;i+1<words.length&&found<200;i++){ const a0=words[i],b0=words[i+1],a=cleanWord(a0,sensitive),b=cleanWord(b0,sensitive); if(!a||!b||a===b) continue; const common=longestCommonSubstring(a,b); const n=Array.from(common).length; if(n<minLen) continue; found++; out.push(`${String(found).padStart(3)}. ${a0} → ${b0}`,`     trasporta: "${common}" (${n} lettere)`); }
  if(!found) out.push('Nessun trasporto consecutivo rilevato.'); if(found>=200) out.push('\n… output limitato a 200 passaggi.'); return out.join('\n');
}
function phraseExists(tokens,parts){ outer:for(let i=0;i+parts.length<=tokens.length;i++){for(let j=0;j<parts.length;j++) if(tokens[i+j]!==parts[j]) continue outer; return true;} return false; }
function findSciarades(text,sensitive,minSeg){
  const raw=tokenizeWords(text), tokens=[], counts=new Map(), display=new Map();
  for(const w of raw){const k=cleanWord(w,sensitive);if(!k)continue;tokens.push(k);counts.set(k,(counts.get(k)||0)+1);if(!display.has(k))display.set(k,w);}
  const targets=[...counts.keys()].filter(k=>Array.from(k).length>=minSeg*2).sort((a,b)=>Array.from(b).length-Array.from(a).length||a.localeCompare(b));
  const out=[`Sciarade interne · segmenti presenti nel testo`,`Segmento minimo: ${minSeg} lettere`,'']; let found=0;
  for(const target of targets){ const rr=Array.from(target), segSet=new Set(), segs=[];
    for(let i=minSeg;i<=rr.length-minSeg;i++){ const a=rr.slice(0,i).join(''),b=rr.slice(i).join(''); if(!counts.has(a)||!counts.has(b)||!phraseExists(tokens,[a,b])) continue; const key=`${a}|${b}`;if(!segSet.has(key)){segSet.add(key);segs.push(key);} }
    for(let i=minSeg;i<=rr.length-2*minSeg;i++) for(let j=i+minSeg;j<=rr.length-minSeg;j++){ const a=rr.slice(0,i).join(''),b=rr.slice(i,j).join(''),c=rr.slice(j).join(''); if(!counts.has(a)||!counts.has(b)||!counts.has(c)||!phraseExists(tokens,[a,b,c]))continue; const key=`${a}|${b}|${c}`;if(!segSet.has(key)){segSet.add(key);segs.push(key);} }
    if(!segs.length) continue; found++; let head=display.get(target); if((counts.get(target)||0)>1) head+=`  [forma intera ×${counts.get(target)}]`; out.push(head);
    segs.sort().forEach(seg=>out.push('  → '+seg.split('|').map(p=>display.get(p)||p).join(' + ')));
    if((counts.get(target)||0)>1) out.push('  ↳ possibile salto di dominio: stessa forma intera + segmentazioni diverse; verifica il senso nel contesto.'); out.push('');
    if(found>=100){out.push('… altri candidati omessi.');break;}
  }
  if(!found) out.push('Nessuna sciarada interna rilevata con il vocabolario del testo.'); return out.join('\n');
}
function repeatedWordsSummary(text,sensitive,maxItems){
  const {freq,display}=freqWords(tokenizeWords(text),sensitive); const a=[...freq.entries()].filter(([,n])=>n>1).sort((x,y)=>y[1]-x[1]||x[0].localeCompare(y[0])).slice(0,maxItems); return a.length?a.map(([k,n])=>`${display.get(k)} = ${n}`).join('\n'):'Nessuna parola ripetuta.';
}
function kuntaPhenomena(text,sensitive){ return limitLines([
  'KUNTA IL TESTO','==============================','','RIPETIZIONI',repeatedWordsSummary(text,sensitive,20),'','SCHEMA RIME / DESINENZE',analyzeRhymeScheme(text,sensitive,3),'','RIMA BACIATA / INCLUSIONE PROGRESSIVA',findInclusionChains(tokenizeWords(text),sensitive,3),'','LETTERTRANSPORT',findLetterTransport(tokenizeWords(text),sensitive,3),'','SCIARADE / SALTI DI DOMINIO',findSciarades(text,sensitive,2)
].join('\n'),900); }

function analyze(text,op,p,sensitive){
  const words=tokenizeWords(text), lines=normalizeLines(text);
  switch(op){
    case 0:{ if(!p.trim())return 'Scrivi nel campo Parametro le lettere da contare.'; const seen=new Set(),targets=[]; for(const r of Array.from(p)){if(/\s/u.test(r))continue;const k=normalizeCase(r,sensitive);if(!seen.has(k)){seen.add(k);targets.push(k);}} const counts=new Map(targets.map(x=>[x,0])); for(const r of Array.from(normalizeCase(text,sensitive))) if(seen.has(r))counts.set(r,(counts.get(r)||0)+1); const total=[...counts.values()].reduce((a,b)=>a+b,0); return targets.map(r=>`${r} = ${counts.get(r)||0}`).join('\n')+`\n\nTotale caratteri cercati: ${total}`; }
    case 1: return `Caratteri (Unicode): ${Array.from(text).length}\nByte UTF-8: ${new TextEncoder().encode(text).length}`;
    case 2: return `Parole: ${words.length}\nParole distinte: ${distinctWordCount(words,sensitive)}`;
    case 3: return `Righe: ${lines.length}\nRighe non vuote: ${lines.filter(x=>x.trim()).length}`;
    case 4:{let spaces=0,white=0;for(const r of Array.from(text)){if(r===' ')spaces++;if(/\s/u.test(r))white++;}return `Spazi semplici: ${spaces}\nCaratteri bianchi totali: ${white}`;}
    case 5: return `Cifre: ${(text.match(/\p{N}/gu)||[]).length}`;
    case 6: return `Segni di punteggiatura: ${(text.match(/\p{P}/gu)||[]).length}`;
    case 7:{const q=p.trim();if(!q)return 'Scrivi una parola o stringa da cercare.';return `"${q}"\nOccorrenze: ${countOccurrences(normalizeCase(text,sensitive),normalizeCase(q,sensitive))}`;}
    case 8:{const {freq,display}=freqWords(words,sensitive);const keys=[...freq.keys()].sort();return limitLines(`Parole uniche: ${keys.length}\n\n`+keys.map(k=>display.get(k)).join('\n'),700);}
    case 9:{const {freq,display}=freqWords(words,sensitive);const a=[...freq.entries()].filter(([,n])=>n>1).sort((x,y)=>y[1]-x[1]||x[0].localeCompare(y[0]));return limitLines(`Parole ripetute: ${a.length}\n\n`+a.map(([k,n])=>`${display.get(k)} = ${n}`).join('\n'),700);}
    case 10:{const f=new Map();for(const r of Array.from(normalizeCase(text,sensitive)))if(/\p{L}/u.test(r))f.set(r,(f.get(r)||0)+1);return [...f.entries()].sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0])).map(([r,n])=>`${r} = ${n}`).join('\n');}
    case 11:{const {freq,display}=freqWords(words,sensitive);return limitLines([...freq.entries()].sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0])).map(([k,n])=>`${display.get(k)} = ${n}`).join('\n'),700);}
    case 12:{if(!words.length)return 'Nessuna parola.';let min=1e9,max=0,mins=[],maxs=[];for(const w of words){const n=Array.from(cleanWord(w,true)).length;if(!n)continue;if(n<min){min=n;mins=[w]}else if(n===min&&!mins.includes(w))mins.push(w);if(n>max){max=n;maxs=[w]}else if(n===max&&!maxs.includes(w))maxs.push(w);}return `Più corta (${min}): ${mins.join(', ')}\n\nPiù lunga (${max}): ${maxs.join(', ')}`;}
    case 13:return findPalindromes(words,sensitive,parseMinLen(p));
    case 14:return findBifronti(words,sensitive,parseMinLen(p));
    case 15:return findAnagrams(words,sensitive,parseMinLen(p));
    case 16:{const a=[];for(const line of lines){const t=line.trim();if(!t)continue;const m=t.match(/[\p{L}\p{N}]/u);if(m)a.push(m[0]);}return `Acrostico (${a.length} righe):\n\n${a.join('')}`;}
    case 17:{const a=[];for(const line of lines){const t=line.trim();if(!t)continue;const rr=Array.from(t);for(let i=rr.length-1;i>=0;i--)if(/[\p{L}\p{N}]/u.test(rr[i])){a.push(rr[i]);break;}}return `Telestico (${a.length} righe):\n\n${a.join('')}`;}
    case 18:return reverseRunes(text);
    case 19:return [...words].reverse().join(' ');
    case 20:return [...words].sort((a,b)=>normalizeCase(a,sensitive).localeCompare(normalizeCase(b,sensitive),'it')).join('\n');
    case 21:{const nums=text.match(/\p{N}+/gu)||[];return `Sequenze numeriche: ${nums.length}\n\n${nums.join('\n')}`;}
    case 22:return analyzeRhymeScheme(text,sensitive,parsePositive(p,3,1,12));
    case 23:return findInclusionChains(words,sensitive,parsePositive(p,3,2,20));
    case 24:return findLetterTransport(words,sensitive,parsePositive(p,3,2,20));
    case 25:return findSciarades(text,sensitive,parsePositive(p,2,1,8));
    case 26:return kuntaPhenomena(text,sensitive);
    default:return 'Operazione non riconosciuta.';
  }
}

function run(){
  if(!input.value){ output.value='Non c’è testo da analizzare.'; return; }
  output.value=analyze(input.value,Number(opSel.value),param.value,$('caseSensitive').checked);
  output.scrollTop=0;
}
$('runBtn').addEventListener('click',run);
$('.quick');
document.querySelectorAll('button.quick').forEach(b=>b.addEventListener('click',()=>{opSel.value=b.dataset.op;updateHint();run();}));
$('clearBtn').addEventListener('click',()=>{input.value='';output.value='';param.value='';input.focus();});
$('openBtn').addEventListener('click',()=>$('fileInput').click());
$('fileInput').addEventListener('change',async e=>{const f=e.target.files[0];if(!f)return;input.value=await f.text();e.target.value='';});
$('pasteBtn').addEventListener('click',async()=>{try{input.value=await navigator.clipboard.readText();}catch{input.focus();$('status').textContent='Clipboard non autorizzata: tieni premuto e usa Incolla.';}});
$('copyBtn').addEventListener('click',async()=>{if(!output.value)return;try{await navigator.clipboard.writeText(output.value);$('status').textContent='Risultato copiato.';}catch{output.select();document.execCommand('copy');}});
$('shareBtn').addEventListener('click',async()=>{if(!output.value)return;if(navigator.share){try{await navigator.share({title:'Kunta',text:output.value});}catch{}}else{$('status').textContent='Condivisione di sistema non disponibile in questo browser.';}});

const shared=localStorage.getItem('kunta.sharedText');
if(shared){input.value=shared;localStorage.removeItem('kunta.sharedText');$('status').textContent='Testo ricevuto dalla condivisione.';}

let installPrompt=null;
window.addEventListener('beforeinstallprompt',e=>{e.preventDefault();installPrompt=e;$('installBtn').classList.remove('hidden');});
$('installBtn').addEventListener('click',async()=>{if(!installPrompt)return;installPrompt.prompt();await installPrompt.userChoice;installPrompt=null;$('installBtn').classList.add('hidden');});
window.addEventListener('appinstalled',()=>{$('status').textContent='Kunta installato nella schermata Home.';});

if('serviceWorker' in navigator && location.protocol.startsWith('http')) navigator.serviceWorker.register('./sw.js');
